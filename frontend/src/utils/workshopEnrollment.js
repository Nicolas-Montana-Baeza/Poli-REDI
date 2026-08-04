export const hasConfirmedWorkshopEnrollment = (
  workshop,
  enrollments = []
) => {
  if (workshop?.isEnrolled === true) {
    return true
  }

  const workshopId = Number(workshop?.id)

  if (!Number.isInteger(workshopId) || workshopId <= 0) {
    return false
  }

  return enrollments.some((enrollment) => (
    Number(enrollment.workshopId) === workshopId &&
    String(enrollment.status || '').toUpperCase() === 'CONFIRMED'
  ))
}

export const isWorkshopAvailabilityItem = (item) => Boolean(
  item?.isWorkshop === true && item?.workshop
)

export const getWorkshopEnrollmentLabel = (item) => (
  item?.workshop?.isEnrolled === true ? 'Inscrito' : ''
)
