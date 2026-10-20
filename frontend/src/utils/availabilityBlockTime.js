// Los horarios administrativos se interpretan en America/Santiago,
// independientemente de la zona horaria configurada en el dispositivo.
const ZONE = 'America/Santiago'

const wallClockFormatter = new Intl.DateTimeFormat('en-GB', {
  timeZone: ZONE,
  year: 'numeric', month: '2-digit', day: '2-digit',
  hour: '2-digit', minute: '2-digit', hourCycle: 'h23'
})

const offsetFormatter = new Intl.DateTimeFormat('en-US', {
  timeZone: ZONE, timeZoneName: 'longOffset'
})

const partsOf = (formatter, value) => Object.fromEntries(
  formatter.formatToParts(value)
    .filter(part => part.type !== 'literal')
    .map(part => [part.type, part.value])
)

export const chileDateKey = (date = new Date()) => {
  const parts = partsOf(wallClockFormatter, date)
  return `${parts.year}-${parts.month}-${parts.day}`
}

export const shiftDateKey = (value, days) => {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) {
    throw new Error('Fecha inválida.')
  }
  const date = new Date(`${value}T12:00:00.000Z`)
  if (Number.isNaN(date.getTime()) || date.toISOString().slice(0, 10) !== value) {
    throw new Error('Fecha inválida.')
  }
  date.setUTCDate(date.getUTCDate() + days)
  return date.toISOString().slice(0, 10)
}

const offsetInMinutes = (date) => {
  const name = partsOf(offsetFormatter, date).timeZoneName
  if (name === 'GMT') return 0
  const match = /^GMT([+-])(\d{2}):(\d{2})$/.exec(name || '')
  if (!match) throw new Error('No se pudo determinar el horario de Santiago.')
  return (match[1] === '+' ? 1 : -1) *
    (Number(match[2]) * 60 + Number(match[3]))
}

export const santiagoLocalToISOString = (value) => {
  const match = /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})$/.exec(value || '')
  if (!match) throw new Error('Indica una fecha y hora válidas.')
  const [, y, mo, d, h, mi] = match.map((part, index) => index ? Number(part) : part)
  const utcWall = Date.UTC(y, mo - 1, d, h, mi)
  const date = new Date(utcWall)
  if (date.getUTCFullYear() !== y || date.getUTCMonth() + 1 !== mo ||
      date.getUTCDate() !== d || date.getUTCHours() !== h ||
      date.getUTCMinutes() !== mi) {
    throw new Error('La fecha u hora indicada no existe.')
  }
  const offsets = new Set([-36, 0, 36].map(hours =>
    offsetInMinutes(new Date(utcWall + hours * 3600000))
  ))
  const valid = [...offsets]
    .map(offset => new Date(utcWall - offset * 60000))
    .filter(candidate => {
      const p = partsOf(wallClockFormatter, candidate)
      return Number(p.year) === y && Number(p.month) === mo &&
        Number(p.day) === d && Number(p.hour) === h && Number(p.minute) === mi
    })
  if (!valid.length) throw new Error('La hora no existe por un cambio de horario en Chile.')
  if (valid.length > 1) throw new Error('Hora ambigua por cambio de horario. Elige otra hora.')
  return valid[0].toISOString()
}

export const formatChileBlockTime = (value) => new Intl.DateTimeFormat('es-CL', {
  timeZone: ZONE, day: '2-digit', month: 'short', year: 'numeric',
  hour: '2-digit', minute: '2-digit', hourCycle: 'h23'
}).format(new Date(value))
