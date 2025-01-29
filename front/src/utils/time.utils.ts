import { format, parse } from 'date-fns'

const dateFormat = 'yyyy-MM-dd'
const timeFormat = 'yyyy-MM-dd HH:mm'

const dateInString = (date: Date) => {
  const dateNew = new Date(date)
  return format(dateNew, dateFormat)
}

const stringInDate = (string: string) => {
  return parse(string, dateFormat, new Date())
}

const timeInString = (date: Date) => {
  const dateNew = new Date(date)
  return format(dateNew, timeFormat)
}

const stringInTime = (string: string) => {
  return parse(string, timeFormat, new Date())
}

const timeUtils = {
  dateInString,
  stringInDate,
  timeInString,
  stringInTime,
  dateFormat,
  timeFormat,
}

export default timeUtils
