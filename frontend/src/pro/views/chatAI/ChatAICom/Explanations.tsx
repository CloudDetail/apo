import { Image } from 'antd'
import mock1Png from 'src/pro/assets/1.png'
import mock2Png from 'src/pro/assets/2.png'
import mock5Png from 'src/pro/assets/5.png'
import mock61Png from 'src/pro/assets/6-1.png'
import mock62Png from 'src/pro/assets/6-2.png'
import mock7Png from 'src/pro/assets/7.png'
import mock8Png from 'src/pro/assets/8.png'

export const Explanations = [
  <div>
    <div>拓扑示例如下：通过告警深入分析可以看到告警关联的完整业务调用关系。</div>
    <Image src={mock1Png} width={'60%'} />
  </div>,
  <>
    <div>
      示例：在告警深入分析中，可以通过拖动时间范围可以看到不同时间段的告警详细信息。智能对话中，提交的是这段时间所有的告警信息，并且每个告警进行已经被关联至应用接口
    </div>
    <Image src={mock2Png} width={'90%'} />
  </>,
  <>先请LLM根据拓扑结构和告警信息完成疑似根因节点的识别。</>,
  <>
    告警信息详细解释：
    <ol style={{ listStyleType: 'disc', paddingLeft: 20 }}>
      <li>应用层告警（URL接口延时异常和错误异常）</li>
      <li>容器告警（k8s事件比如容器killed， 容器重启等异常）</li>
      <li>基础设施主机层告警（主机的CPU、 内存、磁盘、网卡吞吐量异常）</li>
      <li>网络层告警（节点之间连接的网络周期性Ping值异常）</li>
      <li>
        错误告警（Java程序使用java
        Exception，非Java程序，使用日志错误数异常。解决微服务场景中会识别不了错误的场景、将错误封装到error
        code中，导致http status code永远是200）
      </li>
    </ol>
    专家经验：只有应用层告警和错误告警能够在调用链路上传递。
  </>,
  <>
    <div>
      <div className="w-full">
        <div>如下图示例，应用层告警和错误告警都是从下游往上游传递。</div>
        <div>上游节点的应用告警是由下游传递导致的，所以上游节点不是疑似根因节点</div>
        <div>
          {`示例中应用告警传递过程 ts-route-service -> ts-travel-service -> ts-route-plan-service
                    -> ts-gateway-servcie。`}
        </div>
      </div>
      <Image src={mock5Png} width={'100%'} />
    </div>
  </>,
  <>
    <div>如果应用接口产生的告警是偶发的，应该排除出疑似根因节点以外。</div>
    <div>偶发的告警定义：在故障排查时间段内，告警新增之后，很快的看到了已经解决。</div>
    <div>新增发生在：11：19：35秒</div>
    <Image src={mock61Png} width={'100%'} />

    <div>已解决发生在：11：20：35秒</div>
    <Image src={mock62Png} width={'100%'} />
  </>,
  <>
    <div>根因节点产生的告警，应该沿着调用链路逆向传递到业务入口。</div>
    <div>示例中，告警只传递到了某个上游就断掉了，说明该上游和下游告警节点都不是根因节点。</div>
    <Image src={mock7Png} width={'100%'} />
  </>,
  <>
    <div>汇总</div>
    <div>不同大模型的准确率不同，apo 使用 deepseek 模型进行分析</div>
    <div>用户可以修改配置文件选择不同的大模型，如智谱、千帆、豆包等模型</div>
    <Image src={mock8Png} width={'100%'} />
  </>,
]
