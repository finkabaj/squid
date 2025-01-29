import { CSSProperties, PropsWithChildren, useContext } from 'react'
import ValidationContext from '../../Validation/ValidationContext.ts'
import { DatePickerProps, DatePicker, ConfigProvider } from 'antd'
import ru_RU from 'antd/locale/ru_RU'
import dayjs from 'dayjs'
import customParseFormat from 'dayjs/plugin/customParseFormat'
import timeUtils from '../../../utils/time.utils.ts'
import { FaCalendarAlt } from "react-icons/fa";
import CalendarIcon from "../../../assets/icons/calendar.svg?react"



export interface IDateProps {
  style?: CSSProperties
  value?: Date
  name: string
  onChange: (val: Date, name: string) => void
  className?: string
  placeholder?: string
  label?: string
  serverError?: string
  error?: boolean
  errorText?: string
  allowClear?: boolean
  showTime?: boolean
  minDate?: Date
  maxDate?: Date
}

const DateInput = ({
  onChange,
  value,
  children,
  className,
  style,
  serverError,
  placeholder,
  showTime = false,
  minDate = new Date(1910, 0, 1),
  maxDate = new Date(2009, 0, 1),
  ...props
}: PropsWithChildren<IDateProps>) => {
  const context = useContext(ValidationContext)
  const error = context.errors[props.name]?.message || props.error
  const isError = Boolean(error || serverError)

  dayjs.extend(customParseFormat)

  const handleChange: DatePickerProps['onChange'] = (_, dateString) => {
    if (showTime) {
      onChange(timeUtils.stringInTime(dateString as string), props.name)
    } else {
      onChange(timeUtils.stringInDate(dateString as string), props.name)
    }
  }

  const formattedValue =
    value && showTime
      ? dayjs(timeUtils.timeInString(value), timeUtils.timeFormat)
      : value && !showTime
        ? dayjs(timeUtils.dateInString(value), timeUtils.dateFormat)
        : undefined

  return (
    <div className='input-wrapper'>
      {children}
      <ConfigProvider locale={ru_RU}>
        <DatePicker
          showTime={showTime}
          format={showTime ? 'YYYY-MM-DD HH:mm' : 'YYYY-MM-DD'}
          className='date-picker'
          minuteStep={5}
          showNow={false}
          placeholder={placeholder}
          onChange={handleChange}
          style={{
            fontSize: '16px',
            color: '#2F3D53',
          }}
          value={formattedValue}
          suffixIcon={<CalendarIcon className='calendar-icon' />}
          allowClear={true}
          needConfirm={false}
          status={isError ? 'error' : ''}
          minDate={dayjs(minDate)}
          maxDate={dayjs(maxDate)}
          {...props}
        />
      </ConfigProvider>
      {isError && <span className={'text-error'}>{error}</span>}
    </div>
  )
}

export default DateInput
