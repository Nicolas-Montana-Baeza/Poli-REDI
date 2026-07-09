const dayNames = [
  'domingo',
  'lunes',
  'martes',
  'miercoles',
  'jueves',
  'viernes',
  'sabado'
]

const normalize = (value) => {
  return String(value || '')
    .replace(/Ã¡/g, 'a')
    .replace(/Ã©/g, 'e')
    .replace(/Ã­/g, 'i')
    .replace(/Ã³/g, 'o')
    .replace(/Ãº/g, 'u')
    .replace(/Ã±/g, 'n')
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .toLowerCase()
}

const getDateDayName = (dateKey) => {
  const date = new Date(`${dateKey}T00:00:00`)

  if (Number.isNaN(date.getTime())) {
    return ''
  }

  return dayNames[date.getDay()]
}

const getMinutes = (time) => {
  const [hour, minute] = String(time).split(':').map(Number)

  return hour * 60 + minute
}

const getDurationMinutes = (start, end) => {
  return Math.max(0, getMinutes(end) - getMinutes(start))
}

const dayMatchesRange = (text, selectedDay) => {
  const rangeMatch =
    text.match(/(lunes|martes|miercoles|jueves|viernes|sabado|domingo)\s+a\s+(lunes|martes|miercoles|jueves|viernes|sabado|domingo)/)

  if (!rangeMatch) {
    return false
  }

  const startIndex = dayNames.indexOf(rangeMatch[1])
  const endIndex = dayNames.indexOf(rangeMatch[2])
  const selectedIndex = dayNames.indexOf(selectedDay)

  if (startIndex < 0 || endIndex < 0 || selectedIndex < 0) {
    return false
  }

  if (startIndex <= endIndex) {
    return selectedIndex >= startIndex && selectedIndex <= endIndex
  }

  return selectedIndex >= startIndex || selectedIndex <= endIndex
}

export const workshopOccursOnDate = (workshop, selectedDate) => {
  const selectedDay = getDateDayName(selectedDate)
  const dayText = normalize(workshop.dayText)

  if (!selectedDay || !dayText) {
    return false
  }

  return (
    dayText.includes(selectedDay) ||
    dayMatchesRange(dayText, selectedDay)
  )
}

export const getWorkshopTimeRangesForDate = (
  workshop,
  selectedDate
) => {
  const selectedDay = getDateDayName(selectedDate)
  const scheduleText = normalize(workshop.scheduleText)
  const ranges = []
  const pattern =
    /(?:(lunes|martes|miercoles|jueves|viernes|sabado|domingo)\s+)?(\d{1,2}:\d{2})\s*a\s*(\d{1,2}:\d{2})/g

  let match = pattern.exec(scheduleText)

  while (match) {
    const [, explicitDay, start, end] = match

    if (!explicitDay || explicitDay === selectedDay) {
      ranges.push({
        start,
        end,
        durationMinutes: getDurationMinutes(start, end)
      })
    }

    match = pattern.exec(scheduleText)
  }

  return ranges.filter(range => range.durationMinutes > 0)
}

export const buildWorkshopAvailabilityItems = ({
  workshops,
  resources,
  selectedDate
}) => {
  return workshops.flatMap((workshop) => {
    if (
      !workshop.isActive ||
      !workshop.resourceId ||
      !workshopOccursOnDate(workshop, selectedDate)
    ) {
      return []
    }

    const resource = resources.find(
      item => item.id === workshop.resourceId
    )

    return getWorkshopTimeRangesForDate(
      workshop,
      selectedDate
    ).map((range, index) => ({
      id: `workshop-${workshop.id}-${selectedDate}-${index}`,
      resourceId: workshop.resourceId,
      startTime: `${selectedDate}T${range.start}:00`,
      durationMinutes: range.durationMinutes,
      status: 'CONFIRMED',
      title: `Taller: ${workshop.title}`,
      type: 'workshop',
      resourceName: resource?.name || workshop.location || 'Recurso',
      userFullName: 'Taller deportivo',
      userEmail: '',
      userRut: '',
      isWorkshop: true,
      workshop
    }))
  })
}
