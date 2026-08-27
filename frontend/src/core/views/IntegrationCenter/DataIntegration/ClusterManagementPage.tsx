import { Alert, Button, Input, Modal, Segmented, Select, Space, Table, Tag, Typography } from 'antd'
import { ArrowLeft, Network, Pencil, Plus, RefreshCw, Server } from 'lucide-react'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  createDataIntegrationApi,
  getAgentConfigNodesApi,
  getIntegrationClusterListApi,
  updateClusterDisplayNameApi,
  updateNodeClusterIdentityApi,
} from 'src/core/api/integration'
import { notify } from 'src/core/utils/notify'
import styles from './clusterManagementPage.module.scss'

type ClusterType = 'k8s' | 'vm'
type Cluster = { id: string; name?: string; clusterType?: string }
type Node = {
  id: string
  name?: string
  clusterId?: string
  environment?: string
  clusterIp?: string
  mixedCluster?: boolean
}
type ClusterRow = {
  key: string
  id: string
  name: string
  clusterType?: ClusterType
  nodeCount: number
  clusterIps: string[]
  mixed: boolean
  registered?: Cluster
}

const normalizeType = (value?: string): ClusterType | undefined => {
  const normalized = String(value || '')
    .trim()
    .toLowerCase()
  if (['k8s', 'kubernetes', 'helm'].includes(normalized)) return 'k8s'
  if (['vm', 'binary'].includes(normalized)) return 'vm'
  return undefined
}

const typeLabel = (value?: ClusterType) =>
  value === 'k8s' ? 'Kubernetes' : value === 'vm' ? 'VM' : '类型未知'

export default function ClusterManagementPage() {
  const navigate = useNavigate()
  const [clusters, setClusters] = useState<Cluster[]>([])
  const [nodes, setNodes] = useState<Node[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [createOpen, setCreateOpen] = useState(false)
  const [newCluster, setNewCluster] = useState<{
    id: string
    name: string
    clusterType: ClusterType
  }>({
    id: '',
    name: '',
    clusterType: 'k8s',
  })
  const [editingCluster, setEditingCluster] = useState<Cluster>()
  const [displayName, setDisplayName] = useState('')
  const [editingNode, setEditingNode] = useState<Node>()
  const [targetClusterId, setTargetClusterId] = useState('')

  const openCreateCluster = () => {
    setNewCluster({ id: '', name: '', clusterType: 'k8s' })
    setCreateOpen(true)
  }

  const closeCreateCluster = () => {
    setCreateOpen(false)
    setNewCluster({ id: '', name: '', clusterType: 'k8s' })
  }

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [clusterResult, nodeResult] = await Promise.all([
        getIntegrationClusterListApi(),
        getAgentConfigNodesApi(),
      ])
      setClusters(Array.isArray(clusterResult) ? clusterResult : [])
      setNodes(Array.isArray((nodeResult as any)?.nodes) ? (nodeResult as any).nodes : [])
    } catch (error: any) {
      console.error('[ClusterManagementPage] failed to load cluster ownership', error)
      notify({ type: 'error', message: error?.message || '加载集群归属失败' })
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const clusterRows = useMemo<ClusterRow[]>(() => {
    const ids = new Set([
      ...clusters.map((cluster) => cluster.id),
      ...nodes.map((node) => node.clusterId || ''),
    ])
    ids.delete('')
    return [...ids]
      .map((id) => {
        const registered = clusters.find((cluster) => cluster.id === id)
        const members = nodes.filter((node) => node.clusterId === id)
        const memberTypes = new Set(
          members.map((node) => normalizeType(node.environment)).filter(Boolean),
        )
        const clusterType =
          normalizeType(registered?.clusterType) ||
          (memberTypes.size === 1 ? [...memberTypes][0] : undefined)
        return {
          key: id,
          id,
          name: registered?.name || id,
          clusterType,
          nodeCount: members.length,
          clusterIps: [
            ...new Set(members.map((node) => node.clusterIp).filter(Boolean) as string[]),
          ].sort(),
          mixed: members.some((node) => node.mixedCluster) || memberTypes.size > 1,
          registered,
        }
      })
      .sort((left, right) => left.id.localeCompare(right.id))
  }, [clusters, nodes])

  const createCluster = async () => {
    const id = newCluster.id.trim()
    if (!id) return
    setSaving(true)
    try {
      await createDataIntegrationApi({
        id,
        name: newCluster.name.trim() || id,
        clusterType: newCluster.clusterType,
        isMinimal: true,
      })
      setCreateOpen(false)
      setNewCluster({ id: '', name: '', clusterType: 'k8s' })
      notify({ type: 'success', message: `集群 ${id} 已创建` })
      await load()
    } catch (error: any) {
      console.error('[ClusterManagementPage] failed to create cluster', error)
      notify({ type: 'error', message: error?.message || '创建集群失败' })
    } finally {
      setSaving(false)
    }
  }

  const saveDisplayName = async () => {
    if (!editingCluster) return
    setSaving(true)
    try {
      await updateClusterDisplayNameApi(editingCluster.id, displayName.trim())
      setEditingCluster(undefined)
      notify({ type: 'success', message: '集群展示名已更新' })
      await load()
    } catch (error: any) {
      console.error('[ClusterManagementPage] failed to update display name', error)
      notify({ type: 'error', message: error?.message || '更新集群展示名失败' })
    } finally {
      setSaving(false)
    }
  }

  const openNodeEditor = (node: Node) => {
    setEditingNode(node)
    setTargetClusterId(node.clusterId || '')
  }

  const saveNodeCluster = async () => {
    if (!editingNode || !targetClusterId) return
    setSaving(true)
    try {
      await updateNodeClusterIdentityApi(editingNode.id, {
        clusterId: targetClusterId,
      })
      setEditingNode(undefined)
      notify({
        type: 'success',
        message: `节点 ${editingNode.name || editingNode.id} 的归属已更新`,
      })
      await load()
    } catch (error: any) {
      console.error('[ClusterManagementPage] failed to update node ownership', error)
      notify({ type: 'error', message: error?.message || '更新节点归属失败' })
    } finally {
      setSaving(false)
    }
  }

  const targetOptions = clusters
    .filter(
      (cluster) => normalizeType(cluster.clusterType) === normalizeType(editingNode?.environment),
    )
    .map((cluster) => ({
      value: cluster.id,
      label: `${cluster.name || cluster.id} (${cluster.id})`,
    }))

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <div className={styles.titleBlock}>
          <Button
            type="text"
            aria-label="返回探针管理"
            icon={<ArrowLeft size={17} />}
            onClick={() => navigate('/probe-management')}
          />
          <div>
            <Typography.Title level={3}>集群管理</Typography.Title>
            <Typography.Text type="secondary">登记集群身份并调整节点归属</Typography.Text>
          </div>
        </div>
        <Space>
          <Button icon={<RefreshCw size={15} />} loading={loading} onClick={load}>
            刷新
          </Button>
          <Button type="primary" icon={<Plus size={15} />} onClick={openCreateCluster}>
            创建集群
          </Button>
        </Space>
      </div>

      <Alert
        type="warning"
        showIcon
        message="节点归属会改变配置作用范围"
        description="集群与节点的 Kubernetes/VM 类型不可修改，节点只能迁移到同类型集群。相同 cluster_ip 同时出现两种节点类型时，该范围会暂停页面配置下发，默认配置仍按节点类型生效。"
      />

      <section className={styles.section}>
        <div className={styles.sectionHeading}>
          <div>
            <Network size={17} />
            <Typography.Title level={5}>集群</Typography.Title>
          </div>
          <Typography.Text type="secondary">{clusterRows.length} 个 cluster_id</Typography.Text>
        </div>
        <Table<ClusterRow>
          size="small"
          loading={loading}
          rowKey="key"
          pagination={false}
          scroll={{ x: 820 }}
          dataSource={clusterRows}
          columns={[
            {
              title: '展示名',
              dataIndex: 'name',
              render: (value) => <Typography.Text strong>{value}</Typography.Text>,
            },
            { title: 'cluster_id', dataIndex: 'id' },
            {
              title: '固有类型',
              dataIndex: 'clusterType',
              width: 130,
              render: (value) => <Tag>{typeLabel(value)}</Tag>,
            },
            { title: '节点', dataIndex: 'nodeCount', width: 90, render: (value) => `${value} 个` },
            {
              title: 'cluster_ip',
              dataIndex: 'clusterIps',
              render: (values: string[]) => (values.length ? values.join(', ') : '-'),
            },
            {
              title: '状态',
              dataIndex: 'mixed',
              width: 120,
              render: (mixed, record) =>
                mixed ? (
                  <Tag color="error">混合集群</Tag>
                ) : record.registered ? (
                  <Tag color="success">已登记</Tag>
                ) : (
                  <Tag color="warning">待登记</Tag>
                ),
            },
            {
              title: '操作',
              width: 120,
              render: (_, record) =>
                record.registered ? (
                  <Button
                    type="link"
                    size="small"
                    icon={<Pencil size={14} />}
                    onClick={() => {
                      setEditingCluster(record.registered)
                      setDisplayName(record.name)
                    }}
                  >
                    修改展示名
                  </Button>
                ) : (
                  <Button
                    type="link"
                    size="small"
                    onClick={() => {
                      setNewCluster({
                        id: record.id,
                        name: record.name === record.id ? '' : record.name,
                        clusterType: record.clusterType || 'k8s',
                      })
                      setCreateOpen(true)
                    }}
                  >
                    登记集群
                  </Button>
                ),
            },
          ]}
        />
      </section>

      <section className={styles.section}>
        <div className={styles.sectionHeading}>
          <div>
            <Server size={17} />
            <Typography.Title level={5}>节点归属</Typography.Title>
          </div>
          <Typography.Text type="secondary">{nodes.length} 个节点</Typography.Text>
        </div>
        <Table<Node>
          size="small"
          loading={loading}
          rowKey="id"
          pagination={{ pageSize: 20, showSizeChanger: true }}
          scroll={{ x: 820 }}
          dataSource={nodes}
          columns={[
            {
              title: '节点',
              render: (_, record) => (
                <div className={styles.nodeIdentity}>
                  <Typography.Text strong>{record.name || record.id}</Typography.Text>
                  <span>{record.id}</span>
                </div>
              ),
            },
            {
              title: '固有类型',
              dataIndex: 'environment',
              width: 130,
              render: (value) => <Tag>{typeLabel(normalizeType(value))}</Tag>,
            },
            { title: '当前 cluster_id', dataIndex: 'clusterId' },
            { title: 'cluster_ip', dataIndex: 'clusterIp', render: (value) => value || '-' },
            {
              title: '风险',
              dataIndex: 'mixedCluster',
              width: 120,
              render: (mixed) => (mixed ? <Tag color="error">类型混合</Tag> : '-'),
            },
            {
              title: '操作',
              width: 120,
              render: (_, record) => (
                <Button
                  type="link"
                  size="small"
                  icon={<Pencil size={14} />}
                  onClick={() => openNodeEditor(record)}
                >
                  调整归属
                </Button>
              ),
            },
          ]}
        />
      </section>

      <Modal
        title="创建集群"
        open={createOpen}
        okText="创建"
        cancelText="取消"
        confirmLoading={saving}
        okButtonProps={{ disabled: !newCluster.id.trim() }}
        onOk={createCluster}
        onCancel={closeCreateCluster}
      >
        <div className={styles.formStack}>
          <label>
            <span>cluster_id</span>
            <Input
              value={newCluster.id}
              disabled={clusterRows.some((row) => row.id === newCluster.id && !row.registered)}
              onChange={(event) =>
                setNewCluster((current) => ({ ...current, id: event.target.value }))
              }
            />
          </label>
          <label>
            <span>展示名</span>
            <Input
              value={newCluster.name}
              placeholder="默认使用 cluster_id"
              onChange={(event) =>
                setNewCluster((current) => ({ ...current, name: event.target.value }))
              }
            />
          </label>
          <label>
            <span>固有类型</span>
            <Segmented
              block
              value={newCluster.clusterType}
              options={[
                { label: 'Kubernetes', value: 'k8s' },
                { label: 'VM', value: 'vm' },
              ]}
              onChange={(value) =>
                setNewCluster((current) => ({ ...current, clusterType: value as ClusterType }))
              }
            />
          </label>
        </div>
      </Modal>

      <Modal
        title="修改集群展示名"
        open={Boolean(editingCluster)}
        okText="保存"
        cancelText="取消"
        confirmLoading={saving}
        onOk={saveDisplayName}
        onCancel={() => setEditingCluster(undefined)}
      >
        <div className={styles.formStack}>
          <Typography.Text type="secondary">cluster_id: {editingCluster?.id}</Typography.Text>
          <Input
            value={displayName}
            placeholder={editingCluster?.id}
            onChange={(event) => setDisplayName(event.target.value)}
          />
        </div>
      </Modal>

      <Modal
        title="调整节点归属"
        open={Boolean(editingNode)}
        okText="确认调整"
        cancelText="取消"
        confirmLoading={saving}
        okButtonProps={{ disabled: !targetClusterId || targetClusterId === editingNode?.clusterId }}
        onOk={saveNodeCluster}
        onCancel={() => setEditingNode(undefined)}
      >
        <div className={styles.formStack}>
          <Alert type="warning" showIcon message="此操作会改变节点接收集群配置的范围" />
          <Typography.Text strong>{editingNode?.name || editingNode?.id}</Typography.Text>
          <Typography.Text type="secondary">
            节点固有类型: {typeLabel(normalizeType(editingNode?.environment))}
          </Typography.Text>
          <label>
            <span>目标 cluster_id</span>
            <Select
              showSearch
              value={targetClusterId || undefined}
              placeholder="选择同类型集群"
              options={targetOptions}
              onChange={setTargetClusterId}
            />
          </label>
          {!targetOptions.length && (
            <Typography.Text type="danger">暂无同类型已登记集群，请先创建集群。</Typography.Text>
          )}
        </div>
      </Modal>
    </div>
  )
}
