import LoaderSpinner from '../LoderSpinner'
interface IProps {
  label?: string
}
const LoaderPage = (props: IProps) => {
  return (
    <div>
      {props.label}
      <LoaderSpinner />
    </div>
  )
}

export default LoaderPage
