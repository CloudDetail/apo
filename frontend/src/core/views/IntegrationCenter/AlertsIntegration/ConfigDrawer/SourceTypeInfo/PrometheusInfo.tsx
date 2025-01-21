import { Typography, Image } from 'antd'
import Paragraph from 'antd/es/typography/Paragraph'
import Title from 'antd/es/typography/Title'
import Text from 'antd/es/typography/Text'
import CopyPre from './CopyPre'
const code1 = `global:
    ...
route:
    ...
receivers:
    ...
    - name: apo-collector
      webhook_configs:
        - send_resolved: true
          url: '<告警推送地址>'`
const code2 = `global:
    ...
route:
    receiver: xxx
    continue: false
    routes:
        - receiver: apo-collector
          continue: true`
const code3 = `global:
    ...
route:
    receiver: apo-collector
    continue: false`
const PrometheusInfo = () => {
  return (
    <>
      <Typography>
        <Title level={4}>Prometheus 告警接入</Title>
        <Text>
          在使用Prometheus作为告警数据源时,可以通过 AlertManager
          的Webhook组件将告警事件推送到APO平台。
        </Text>
        <Typography>下面是AlertManager的配置修改说明.</Typography>
        <Title level={5}>1. 添加Webhookt通知渠道</Title>
        <Text>
          需要修改AlertManager实例的配置信息,通常是AlertManager根目录的 alertmanager.yml; 在
          receivers 列表中添加新的webhook配置项,示例如下:
        </Text>
        <CopyPre code={code1} />
        <Title level={5}>2. 将新增的webhook项添加到通知路由中</Title>
        <Text>推荐将新增的webhook通知渠道作为子路由(routes)加入到通知列表,示例如下</Text>
        <CopyPre code={code2} />
        <Text>或者, 如果不使用原有的推送渠道,可以将新增的webhook对象替换根路由</Text>
        <CopyPre code={code3} />
        <Text>注意, AlertManager通知顺序为:</Text>
        <ol>
          <li>先依次通知route.routes下的通知渠道</li>
          <li>再通知route.receiver指定的通知渠道</li>
        </ol>
        <Text>如果先通知的渠道中设置了 `continue: false` 的配置项, 后续通知渠道不会接收到通知</Text>
        <Title level={5}>3. 保存配置文件</Title>
        <Title level={5}>4. 令AlertManager重新加载配置文件</Title>
        可以重启AlertManager, 或发送POST请求到Alertmanger的`/-/reload` 接口, 使更改生效
        <Title level={5}>5. 完成</Title>
      </Typography>
    </>
  )
}
export default PrometheusInfo
