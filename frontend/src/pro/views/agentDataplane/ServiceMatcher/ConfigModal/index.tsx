import { Modal, Form, Button, Spin } from 'antd'
import { useEffect, useState } from 'react'
import { checkServiceNameRuleApi, getServicesApi, saveServiceRuleApi } from 'src/pro/api/dataplane'
import ConditionsFormList from './ConditionsFormList'
import { useSelector } from 'react-redux'
import { selectProcessedTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { useDebounce } from 'react-use'
import { notify } from 'src/core/utils/notify'
import InstanceTag from '../InstanceTag'
import { useTranslation } from 'react-i18next'
import { LoadingOutlined } from '@ant-design/icons'
const ConfigModal = ({
  serviceName,
  open,
  onCancel,
  info,
  onRefresh,
}: {
  serviceName: string
  open: boolean
  onCancel: () => void
  info: any
  onRefresh: () => void
}) => {
  const { t } = useTranslation('pro/dataplane')
  const [form] = Form.useForm()
  const { startTime, endTime } = useSelector(selectProcessedTimeRange)
  const [instances, setInstances] = useState<any[]>([])
  const [loading, setLoading] = useState(false)
  useDebounce(
    () => {
      if (startTime && endTime) {
        getServicesApi({
          startTime: startTime,
          endTime: endTime,
        }).then(() => {
          // 暂时注释掉，因为services变量已被移除
          // setServices(res.results)
        })
      }
    },
    500,
    [],
  )
  const saveForm = () => {
    form.validateFields().then((values) => {
      saveServiceRuleApi(values)
        .then(() => {
          notify({
            type: 'success',
            message: t('saveSuccess'),
          })
          onCancel()
        })
        .catch(() => {
          notify({
            type: 'error',
            message: t('saveFailed'),
          })
        })
        .finally(() => {
          onRefresh()
        })
    })
  }
  useEffect(() => {
    if (open) {
      form.resetFields()
      if (info) {
        let formValue = {
          service: info.serviceName,
          ruleId: info.id,
          conditions: info.conditions?.map((condition: any) => ({
            conditionId: condition.id,
            matchType: condition.matchType,
            key: condition.key,
            value: condition.value,
          })),
        }
        form.setFieldsValue(formValue)
        checkRule(formValue)
      }
    }
  }, [open, info])

  const serviceValue = Form.useWatch('service', form)
  useEffect(() => {
    if (serviceValue) {
      // getServiceValuesRuleApi({
      //   service: serviceValue,
      //   startTime: startTime,
      //   endTime: endTime,
      // }).then((res) => {
      //   console.log(res)
      // })
    }
  }, [serviceValue])
  const checkRule = (values) => {
    checkServiceNameRuleApi(values)
      .then((res) => {
        setInstances(res?.instances || [])
      })
      .finally(() => {
        setLoading(false)
      })
  }
  return (
    <Modal
      open={open}
      onCancel={onCancel}
      title={t('addServiceMatchRule')}
      width={'50%'}
      styles={{ body: { height: '60vh', overflowY: 'auto', overflowX: 'hidden' } }}
      okText={t('save')}
      onOk={saveForm}
      footer={(_, { OkBtn }) => (
        <>
          <Button
            color="primary"
            variant="outlined"
            onClick={() => {
              setLoading(true)
              form
                .validateFields()
                .then((values) => {
                  checkRule(values)
                })
                .catch(() => {
                  setLoading(false)
                })
            }}
          >
            {t('test')}
          </Button>
          <OkBtn />
        </>
      )}
      destroyOnClose
    >
      <div className="flex flex-col gap-2 bg-[var(--ant-color-primary-bg)] p-2 rounded-md mb-3">
        <div className="text-sm text-[var(--ant-color-text-secondary)] flex justify-between">
          <div>{t('serviceInstancesCount', { count: instances.length })}</div>
          {loading && <Spin indicator={<LoadingOutlined spin />} />}
        </div>
        <div className=" max-h-[200px] overflow-y-auto">
          {instances.map((instance) => (
            <InstanceTag
              instanceName={instance.instanceName}
              source="rule"
              key={instance.instanceName}
            />
          ))}
        </div>
      </div>
      <Form
        form={form}
        labelCol={{ span: 4, offset: 0 }}
        wrapperCol={{ span: 18 }}
        colon={false}
        initialValues={{
          service: serviceName,
          conditions: [{ conditionId: null, matchType: 'equals' }],
        }}
      >
        <Form.Item name="ruleId" hidden></Form.Item>
        <Form.Item label={t('serviceName')} name="service">
          {serviceName}
          {/* <Select
            options={services.map((service) => ({ label: service.name, value: service.id }))}
          /> */}
        </Form.Item>
        <Form.Item label={t('matchRule')} required>
          <ConditionsFormList />
        </Form.Item>
      </Form>
    </Modal>
  )
}
export default ConfigModal
