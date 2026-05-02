# Crear estructura base
$base = "frontend"

# Carpetas
$folders = @(
"$base/public",

"$base/src/assets/images",
"$base/src/assets/icons",
"$base/src/assets/styles",

"$base/src/components/layout",
"$base/src/components/dashboard",
"$base/src/components/availability",
"$base/src/components/admin",
"$base/src/components/ui",
"$base/src/components/forms",

"$base/src/views",
"$base/src/router",
"$base/src/services",
"$base/src/store",
"$base/src/composables",
"$base/src/utils"
)

foreach ($folder in $folders) {
    New-Item -ItemType Directory -Path $folder -Force | Out-Null
}

# Archivos
$files = @(

# Assets
"$base/src/assets/styles/main.css",
"$base/src/assets/styles/variables.css",

# Layout
"$base/src/components/layout/AppLayout.vue",
"$base/src/components/layout/Sidebar.vue",
"$base/src/components/layout/SidebarItem.vue",
"$base/src/components/layout/HeaderBar.vue",
"$base/src/components/layout/NotificationBell.vue",
"$base/src/components/layout/UserMenu.vue",

# Dashboard
"$base/src/components/dashboard/FacilityCarousel.vue",
"$base/src/components/dashboard/FacilityCard.vue",
"$base/src/components/dashboard/ReservationsPanel.vue",
"$base/src/components/dashboard/ReservationCard.vue",
"$base/src/components/dashboard/QuickActions.vue",

# Availability
"$base/src/components/availability/AvailabilitySection.vue",
"$base/src/components/availability/CalendarMini.vue",
"$base/src/components/availability/CalendarToolbar.vue",
"$base/src/components/availability/ScheduleGrid.vue",
"$base/src/components/availability/TimeSlot.vue",
"$base/src/components/availability/ResourceColumn.vue",

# Admin
"$base/src/components/admin/AdminSummary.vue",
"$base/src/components/admin/MetricCard.vue",
"$base/src/components/admin/AlertsPanel.vue",
"$base/src/components/admin/AlertItem.vue",
"$base/src/components/admin/AdminPanelLinks.vue",

# UI
"$base/src/components/ui/PrimaryButton.vue",
"$base/src/components/ui/StatusBadge.vue",
"$base/src/components/ui/InfoBanner.vue",
"$base/src/components/ui/ParticipantsProgress.vue",

# Forms
"$base/src/components/forms/ReservationForm.vue",
"$base/src/components/forms/ResourcePicker.vue",
"$base/src/components/forms/DateTimePicker.vue",

# Views
"$base/src/views/DashboardView.vue",
"$base/src/views/AvailabilityView.vue",
"$base/src/views/ReservationsView.vue",
"$base/src/views/ReservationDetailView.vue",
"$base/src/views/AdminView.vue",
"$base/src/views/NotFoundView.vue",

# Core
"$base/src/router/index.js",
"$base/src/services/api.js",
"$base/src/services/reservationService.js",
"$base/src/services/userService.js",
"$base/src/services/authService.js",

"$base/src/store/index.js",
"$base/src/store/reservationStore.js",
"$base/src/store/userStore.js",

"$base/src/composables/useReservations.js",
"$base/src/composables/useAuth.js",
"$base/src/composables/useAvailability.js",

"$base/src/utils/dateUtils.js",
"$base/src/utils/constants.js",
"$base/src/utils/validators.js",

"$base/src/App.vue",
"$base/src/main.js",

# Root
"$base/.env",
"$base/index.html",
"$base/package.json",
"$base/vite.config.js"
)

foreach ($file in $files) {
    New-Item -ItemType File -Path $file -Force | Out-Null
    write-host "📄 Creado: $file"
}

Write-Host "✅ Estructura creada correctamente"