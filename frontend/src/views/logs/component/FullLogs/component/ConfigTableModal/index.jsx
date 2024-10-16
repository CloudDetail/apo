import { Form, Input, Modal, Select, TreeSelect } from 'antd'
import React, { Children, useEffect, useState } from 'react'
import LogRouteRuleFormList from './component/LogRouteRuleFormList'
import {
  addLogOtherTableApi,
  addLogRuleApi,
  getLogOtherTableInfoApi,
  getLogOtherTableListApi,
  getLogRuleApi,
  updateLogRuleApi,
} from 'src/api/logs'
import { showToast } from 'src/utils/toast'
import { useLogsContext } from 'src/contexts/LogsContext'

const ConfigTableModal = ({ modalVisible, closeModal }) => {
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
        title: '接入日志表配置成功',
        color: 'success',
      })
    })
    getLogTableInfo()
    closeModal()
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
      title={'接入日志表配置'}
      open={modalVisible}
      onCancel={closeModal}
      destroyOnClose
      centered
      okText={'保存'}
      cancelText="取消"
      maskClosable={false}
      onOk={saveLogRule}
      width={1000}
      bodyStyle={{ maxHeight: '80vh', overflowY: 'auto', overflowX: 'hidden' }}
    >
      <Form layout={'vertical'} form={form} preserve={false}>
        <Form.Item label="数据源" name="dataBase" required>
          <TreeSelect
            showSearch
            style={{ width: '100%' }}
            dropdownStyle={{ maxHeight: 400, overflow: 'auto' }}
            placeholder="Please select"
            allowClear
            treeDefaultExpandAll
            onChange={handleTreeSelectChange}
            treeData={tables}
            // showCheckedStrategy="SHOW_ALL"
          />
        </Form.Item>
        <Form.Item label="时间解析字段" name="timeField" required>
          <Select options={tableColumns} labelInValue placeholder="选择匹配规则Key" />
        </Form.Item>
      </Form>
    </Modal>
  )
}
export default ConfigTableModal
