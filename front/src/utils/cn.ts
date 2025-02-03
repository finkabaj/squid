type Value = string | number | boolean | undefined | null
type Mapping = Record<string, unknown>
type Argument = Value | Mapping | ArgumentArray
interface ArgumentArray extends Array<Argument> {}

const cn = (...arg: ArgumentArray[]) => {
  if (typeof arg[0] === 'string') return arg.filter(Boolean).join(' ')

  const classes = []
  const classNames = arg[0] as Mapping

  for (const className in classNames) {
    if (classNames[className]) classes.push(`${className}`)
  }

  return classes.join(' ')
}

export default cn
