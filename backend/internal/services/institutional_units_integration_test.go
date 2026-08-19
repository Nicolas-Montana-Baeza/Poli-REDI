package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

// TestInstitutionalUnitAuthorizationIntegration valida la primera capa del
// módulo de Programación Institucional contra PostgreSQL.
//
// El escenario comprueba que:
//
//   - un administrador puede crear unidades institucionales;
//   - un administrador puede asignar MANAGER y MEMBER;
//   - un MANAGER solo administra las unidades que le fueron asignadas;
//   - un MEMBER no obtiene permisos de programación;
//   - un MANAGER no puede crear nuevas unidades institucionales;
//   - un administrador global puede operar sobre cualquier unidad activa.
//
// Este test utiliza exclusivamente la base PostgreSQL temporal de integración.
func TestInstitutionalUnitAuthorizationIntegration(t *testing.T) {
	if os.Getenv("POLIREDI_INTEGRATION") != "1" {
		t.Skip("integration test disabled")
	}

	t.Setenv("MVP_SCOPE", "mvp2")

	database.Close()

	if err := database.Connect(); err != nil {
		t.Fatalf("connect postgres: %v", err)
	}

	defer database.Close()

	ctx := context.Background()

	// ------------------------------------------------------------------------
	// Limpieza robusta.
	// ------------------------------------------------------------------------
	//
	// Registramos el defer antes de comenzar a crear datos. Así, si el test
	// falla a mitad del flujo, todos los IDs conocidos serán eliminados y no
	// contaminaremos ejecuciones posteriores.

	var (
		createdUnitIDs []int
		createdUserIDs []int
	)

	defer func() {
		// Las membresías referencian tanto unidades como usuarios, por lo que
		// deben eliminarse antes que cualquiera de esas entidades.
		for _, unitID := range createdUnitIDs {
			if _, err := database.DB.ExecContext(
				context.Background(),
				`
				DELETE FROM institutional_unit_memberships
				WHERE unit_id = $1
				`,
				unitID,
			); err != nil {
				t.Logf(
					"cleanup memberships for unit %d: %v",
					unitID,
					err,
				)
			}
		}

		for _, unitID := range createdUnitIDs {
			if _, err := database.DB.ExecContext(
				context.Background(),
				`
				DELETE FROM institutional_units
				WHERE id = $1
				`,
				unitID,
			); err != nil {
				t.Logf(
					"cleanup institutional unit %d: %v",
					unitID,
					err,
				)
			}
		}

		for _, userID := range createdUserIDs {
			if _, err := database.DB.ExecContext(
				context.Background(),
				`
				DELETE FROM users
				WHERE id = $1
				`,
				userID,
			); err != nil {
				t.Logf(
					"cleanup user %d: %v",
					userID,
					err,
				)
			}
		}
	}()

	// ------------------------------------------------------------------------
	// Usuarios temporales.
	// ------------------------------------------------------------------------
	//
	// El sufijo evita colisiones incluso si una ejecución anterior terminó
	// abruptamente antes de ejecutar su cleanup.

	suffix := time.Now().UnixNano()

	createUser := func(
		emailPrefix string,
		fullName string,
		isAdmin bool,
	) models.LocalAuthUser {
		t.Helper()

		email := fmt.Sprintf(
			"%s.%d@test.local",
			emailPrefix,
			suffix,
		)

		var userID int

		err := database.DB.QueryRowContext(
			ctx,
			`
			INSERT INTO users (
				email,
				full_name,
				rut,
				is_admin,
				is_blocked
			)
			VALUES (
				$1,
				$2,
				NULL,
				$3,
				false
			)
			RETURNING id
			`,
			email,
			fullName,
			isAdmin,
		).Scan(&userID)

		if err != nil {
			t.Fatalf(
				"create integration user %s: %v",
				fullName,
				err,
			)
		}

		createdUserIDs = append(createdUserIDs, userID)

		return models.LocalAuthUser{
			ID:        userID,
			Email:     email,
			FullName:  fullName,
			IsAdmin:   isAdmin,
			IsBlocked: false,
		}
	}

	admin := createUser(
		"mvp2.institutional.admin",
		"Integration Institutional Admin",
		true,
	)

	manager := createUser(
		"mvp2.institutional.manager",
		"Integration Institutional Manager",
		false,
	)

	member := createUser(
		"mvp2.institutional.member",
		"Integration Institutional Member",
		false,
	)

	// ------------------------------------------------------------------------
	// Creación de unidades.
	// ------------------------------------------------------------------------

	educationUnit, err := CreateInstitutionalUnit(
		admin,
		models.CreateInstitutionalUnitRequest{
			Name: "  Educación Física Integration " +
				fmt.Sprint(suffix),
			Code:     " efi-" + fmt.Sprint(suffix),
			UnitType: models.InstitutionalUnitTypeAcademicProgram,
		},
	)

	if err != nil {
		t.Fatalf("create Educación Física unit: %v", err)
	}

	createdUnitIDs = append(
		createdUnitIDs,
		educationUnit.ID,
	)

	preparationUnit, err := CreateInstitutionalUnit(
		admin,
		models.CreateInstitutionalUnitRequest{
			Name: "Preparación Física Integration " +
				fmt.Sprint(suffix),
			Code:     "pf-" + fmt.Sprint(suffix),
			UnitType: models.InstitutionalUnitTypeAcademicProgram,
		},
	)

	if err != nil {
		t.Fatalf("create Preparación Física unit: %v", err)
	}

	createdUnitIDs = append(
		createdUnitIDs,
		preparationUnit.ID,
	)

	// Verificamos además la normalización realizada por el service.
	expectedEducationCode := fmt.Sprintf(
		"EFI-%d",
		suffix,
	)

	if educationUnit.Code != expectedEducationCode {
		t.Fatalf(
			"expected normalized code %q, got %q",
			expectedEducationCode,
			educationUnit.Code,
		)
	}

	// ------------------------------------------------------------------------
	// Asignación de roles.
	// ------------------------------------------------------------------------

	managerMembership, err := AddInstitutionalUnitMembership(
		admin,
		educationUnit.ID,
		models.AddInstitutionalUnitMembershipRequest{
			UserID: manager.ID,
			Role:   models.InstitutionalMembershipRoleManager,
		},
	)

	if err != nil {
		t.Fatalf("assign manager membership: %v", err)
	}

	if managerMembership.Role !=
		models.InstitutionalMembershipRoleManager {
		t.Fatalf(
			"expected MANAGER role, got %s",
			managerMembership.Role,
		)
	}

	memberMembership, err := AddInstitutionalUnitMembership(
		admin,
		educationUnit.ID,
		models.AddInstitutionalUnitMembershipRequest{
			UserID: member.ID,
			Role:   models.InstitutionalMembershipRoleMember,
		},
	)

	if err != nil {
		t.Fatalf("assign member membership: %v", err)
	}

	if memberMembership.Role !=
		models.InstitutionalMembershipRoleMember {
		t.Fatalf(
			"expected MEMBER role, got %s",
			memberMembership.Role,
		)
	}

	// ------------------------------------------------------------------------
	// MANAGER puede gestionar su propia unidad.
	// ------------------------------------------------------------------------

	if err := EnsureInstitutionalUnitManager(
		manager,
		educationUnit.ID,
	); err != nil {
		t.Fatalf(
			"manager should manage assigned unit: %v",
			err,
		)
	}

	// ------------------------------------------------------------------------
	// MANAGER no hereda acceso a otras unidades.
	// ------------------------------------------------------------------------

	err = EnsureInstitutionalUnitManager(
		manager,
		preparationUnit.ID,
	)

	if !errors.Is(
		err,
		ErrInstitutionalUnauthorized,
	) {
		t.Fatalf(
			"expected unauthorized for unrelated unit, got %v",
			err,
		)
	}

	// ------------------------------------------------------------------------
	// MEMBER no obtiene permisos de programación.
	// ------------------------------------------------------------------------

	err = EnsureInstitutionalUnitManager(
		member,
		educationUnit.ID,
	)

	if !errors.Is(
		err,
		ErrInstitutionalUnauthorized,
	) {
		t.Fatalf(
			"expected MEMBER to be unauthorized, got %v",
			err,
		)
	}

	// ------------------------------------------------------------------------
	// El listado para MANAGER contiene únicamente sus unidades gestionadas.
	// ------------------------------------------------------------------------

	managerUnits, err := GetInstitutionalUnitsForUser(manager)
	if err != nil {
		t.Fatalf("get manager units: %v", err)
	}

	if len(managerUnits) != 1 {
		t.Fatalf(
			"expected exactly 1 managed unit, got %d",
			len(managerUnits),
		)
	}

	if managerUnits[0].ID != educationUnit.ID {
		t.Fatalf(
			"expected managed unit %d, got %d",
			educationUnit.ID,
			managerUnits[0].ID,
		)
	}

	// MEMBER nuevo no administra ninguna unidad.
	memberUnits, err := GetInstitutionalUnitsForUser(member)
	if err != nil {
		t.Fatalf("get member units: %v", err)
	}

	if len(memberUnits) != 0 {
		t.Fatalf(
			"expected MEMBER to manage 0 units, got %d",
			len(memberUnits),
		)
	}

	// ------------------------------------------------------------------------
	// MANAGER puede consultar las membresías de su propia unidad.
	// ------------------------------------------------------------------------

	memberships, err := GetInstitutionalUnitMemberships(
		manager,
		educationUnit.ID,
	)

	if err != nil {
		t.Fatalf(
			"manager get memberships: %v",
			err,
		)
	}

	if len(memberships) != 2 {
		t.Fatalf(
			"expected 2 memberships, got %d",
			len(memberships),
		)
	}

	// ------------------------------------------------------------------------
	// MANAGER no puede crear unidades institucionales.
	// ------------------------------------------------------------------------

	_, err = CreateInstitutionalUnit(
		manager,
		models.CreateInstitutionalUnitRequest{
			Name:     "Unidad prohibida",
			Code:     fmt.Sprintf("DENIED-%d", suffix),
			UnitType: models.InstitutionalUnitTypeOther,
		},
	)

	if !errors.Is(
		err,
		ErrInstitutionalUnauthorized,
	) {
		t.Fatalf(
			"expected manager unit creation to be unauthorized, got %v",
			err,
		)
	}

	// ------------------------------------------------------------------------
	// MANAGER tampoco puede otorgar permisos a otros usuarios.
	// ------------------------------------------------------------------------

	_, err = AddInstitutionalUnitMembership(
		manager,
		educationUnit.ID,
		models.AddInstitutionalUnitMembershipRequest{
			UserID: member.ID,
			Role:   models.InstitutionalMembershipRoleManager,
		},
	)

	if !errors.Is(
		err,
		ErrInstitutionalUnauthorized,
	) {
		t.Fatalf(
			"expected manager membership assignment to be unauthorized, got %v",
			err,
		)
	}

	// ------------------------------------------------------------------------
	// El administrador global puede actuar sobre cualquier unidad activa.
	// ------------------------------------------------------------------------

	if err := EnsureInstitutionalUnitManager(
		admin,
		preparationUnit.ID,
	); err != nil {
		t.Fatalf(
			"admin should manage any active unit: %v",
			err,
		)
	}

	t.Logf(
		"PASS: admin=%d manager=%d member=%d; units=%d,%d; authorization boundaries preserved",
		admin.ID,
		manager.ID,
		member.ID,
		educationUnit.ID,
		preparationUnit.ID,
	)
}
