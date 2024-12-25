import { Form, Modal, Select, TreeSelect } from 'antd'
import React, { useEffect, useState } from 'react'
import {
  addLogOtherTableApi,
  getLogOtherTableInfoApi,
  getLogOtherTableListApi,
} from 'core/api/logs'
import { showToast } from 'src/core/utils/toast'
import { useLogsContext } from 'src/core/contexts/LogsContext'
import { useTranslation } from 'react-i18next' // 引入i18n

const ConfigTableModal = ({ modalVisible, closeModal }) => {
  const { t } = useTranslation('oss/fullLogs') // 使用i18n
  const { getLogTableInfo, updateLoading } = useLogsContext()
  const [form] = Form.useForm()
  const [tables, setTables] = useState([])
  const [tableColumns, setTableColumns] = useState([])
  const getLogOtherTableList = () => {
    getLogOtherTableListApi().then((res) => {
      const tables = res.otherTables?.map((database) => ({
        key: 'database-' + database.dataBase,
        value: 'database-' + database.dataBase,
        title: database.dataBase,
        selectable: false,
        children: database.tables?.map((table) => ({
          key: 'database-' + database.dataBase + '-table-' + table.tableName,
          value: 'database-' + database.dataBase + '-table-' + table.tableName,
          title: table.tableName,
          database: database.dataBase,
          tableName: table.tableName,
        })),
      }))
      setTables(tables)
    })
  }
  const getLogOtherTableInfo = (selectedNode) => {
    getLogOtherTableInfoApi({
      dataBase: selectedNode.database,
      tableName: selectedNode.tableName,
    }).then((res) => {
      setTableColumns(
        res.columns?.map((column) => ({
          label: column.name,
          value: column.name,
        })),
      )
    })
  }
  useEffect(() => {
    if (modalVisible) getLogOtherTableList()
  }, [modalVisible])

  function addOtherTable(params) {
    addLogOtherTableApi(params).then((res) => {
      showToast({
        title: t('ConfigTableModal.configSuccessToast'),
        color: 'success',
      })

      getLogTableInfo()
      closeModal()
    })
  }
  function saveLogRule() {
    form
      .validateFields({})
      .then(() => {
        const formState = form.getFieldsValue(true)
        const params = {
          dataBase: formState.treeSelect.database,
          tableName: formState.treeSelect.tableName,
          timeField: formState.timeField.value,
        }
        addOtherTable(params)
      })
      .catch((error) => console.log(error))
  }
  const handleTreeSelectChange = (value, label, extra) => {
    form.setFieldsValue({ treeSelect: extra.triggerNode.props }) // 设置表单值为节点数据
    getLogOtherTableInfo(extra.triggerNode.props)
  }
  return (
    <Modal
      title={t('ConfigTableModal.modalTitle')}
      open={modalVisible}
      onCancel={closeModal}
      destroyOnClose
      centered
      okText={t('ConfigTableModal.saveText')}
      cancelText={t('ConfigTableModal.cancelText')}
      maskClosable={false}
      onOk={saveLogRule}
      width={1000}
      bodyStyle={{ maxHeight: '80vh', overflowY: 'auto', overflowX: 'hidden' }}
    >
      <Form layout={'vertical'} form={form} preserve={false}>
        <Form.Item label={t('ConfigTableModal.dataSourceLabel')} name="dataBase" required>
          <TreeSelect
            showSearch
            style={{ width: '100%' }}
            dropdownStyle={{ maxHeight: 400, overflow: 'auto' }}
            placeholder={t('ConfigTableModal.dataSourcePlaceholder')}
            allowClear
            treeDefaultExpandAll
            onChange={handleTreeSelectChange}
            treeData={tables}
            // showCheckedStrategy="SHOW_ALL"
          />
        </Form.Item>
        <Form.Item label={t('ConfigTableModal.timeFieldLabel')} name="timeField" required>
          <Select
            options={tableColumns}
            labelInValue
            placeholder={t('ConfigTableModal.timeFieldPlaceholder')}
          />
        </Form.Item>
      </Form>
    </Modal>
  )
}
export default ConfigTableModal
