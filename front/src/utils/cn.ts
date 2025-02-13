type Value = string | number | boolean | undefined | null
type Mapping = Record<string, unknown>
type Argument = Value | Mapping | ArgumentArray
interface ArgumentArray extends Array<Argument> {}

/** Implementation of the "classnames" library method */
const cn = (...arg: ArgumentArray) => {
  if (typeof arg[0] === 'string') return arg.filter(Boolean).join(' ')

  const classes = []

  for (const className in arg[0] as Array<Record<string, boolean>>) {
    if (arg[0][className]) classes.push(`${className}`)
  }

  return classes.join(' ')
}

export default cn
