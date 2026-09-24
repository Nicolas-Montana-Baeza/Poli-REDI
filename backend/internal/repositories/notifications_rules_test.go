package repositories

import "testing"

func TestParticipantNotificationPendingToConfirmed(t *testing.T) {
	notification, ok := participantNotificationForTransition(
		"PENDING",
		"CONFIRMED",
		GroupConditionPending,
		GroupConditionHealthy,
	)

	if !ok {
		t.Fatal("expected notification")
	}

	if notification.Type != "RESERVATION_CONFIRMED" {
		t.Fatalf(
			"expected RESERVATION_CONFIRMED, got %s",
			notification.Type,
		)
	}
}

func TestParticipantNotificationHealthyToAtRisk(t *testing.T) {
	notification, ok := participantNotificationForTransition(
		"CONFIRMED",
		"CONFIRMED",
		GroupConditionHealthy,
		GroupConditionAtRisk,
	)

	if !ok {
		t.Fatal("expected notification")
	}

	if notification.Type != "RESERVATION_AT_RISK" {
		t.Fatalf(
			"expected RESERVATION_AT_RISK, got %s",
			notification.Type,
		)
	}
}

func TestParticipantNotificationAtRiskToHealthy(t *testing.T) {
	notification, ok := participantNotificationForTransition(
		"CONFIRMED",
		"CONFIRMED",
		GroupConditionAtRisk,
		GroupConditionHealthy,
	)

	if !ok {
		t.Fatal("expected notification")
	}

	if notification.Type != "RESERVATION_RECOVERED" {
		t.Fatalf(
			"expected RESERVATION_RECOVERED, got %s",
			notification.Type,
		)
	}
}

func TestParticipantNotificationIgnoresStableCondition(t *testing.T) {
	_, ok := participantNotificationForTransition(
		"CONFIRMED",
		"CONFIRMED",
		GroupConditionHealthy,
		GroupConditionHealthy,
	)

	if ok {
		t.Fatal("stable HEALTHY condition must not create notification")
	}
}

func TestAdministrativeCancellationNotifiesAffectedOwner(t *testing.T) {
	notification, ok := administrativeCancellationNotification(
		true,
		99,
		10,
	)

	if !ok {
		t.Fatal("expected administrative cancellation notification")
	}

	if notification.Type != "RESERVATION_CANCELLED" {
		t.Fatalf(
			"expected RESERVATION_CANCELLED, got %s",
			notification.Type,
		)
	}
}

func TestOwnerCancellationDoesNotNotifyItself(t *testing.T) {
	_, ok := administrativeCancellationNotification(
		true,
		10,
		10,
	)

	if ok {
		t.Fatal("owner cancelling own reservation must not notify itself")
	}
}

func TestNormalUserCancellationDoesNotCreateAdminNotification(
	t *testing.T,
) {
	_, ok := administrativeCancellationNotification(
		false,
		10,
		10,
	)

	if ok {
		t.Fatal("normal user cancellation must not create admin notification")
	}
}

func TestInstitutionalCancellationNotificationType(t *testing.T) {
	notification := institutionalCancellationNotification()

	if notification.Type !=
		"RESERVATION_CANCELLED_INSTITUTIONAL" {

		t.Fatalf(
			"unexpected type %s",
			notification.Type,
		)
	}
}
