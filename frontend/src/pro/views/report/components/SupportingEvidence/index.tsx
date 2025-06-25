import { Collapse } from 'antd'
import { SlArrowDown } from 'react-icons/sl'
import Logs from './Logs'
import Metrics from './Metrics'
import AffectedEndpoints from '../AffectedEndpoints'

const SupportingEvidence = ({ evidence, startTime, endTime }) => {
  const getChildren = (type, data) => {
    return type === 'log' ? (
      <Logs logs={data} />
    ) : type === 'metrics' ? (
      <Metrics data={data} startTime={startTime} endTime={endTime} />
    ) : (
      <AffectedEndpoints topology={data} />
    )
  }
  return (
    <>
      {evidence.map((item, index) => (
        <div className="mt-3 border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
          <Collapse
            items={[
              {
                key: index,
                label: item.name,
                children: getChildren(item.type, item.data),
              },
            ]}
            defaultActiveKey={'brief'}
            //   className={styles.collapse}
            // bordered={false}
            expandIconPosition={'end'}
            expandIcon={({ isActive }) => {
              return isActive ? (
                <SlArrowDown className={'rotate-180'} size={14} />
              ) : (
                <SlArrowDown className={'rotate-0'} size={14} />
              )
            }}
          />
        </div>
      ))}
    </>
  )
}
export default SupportingEvidence
