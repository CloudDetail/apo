/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { useEffect, useState } from 'react'
import {
  getClusterInstallCmdApi,
  getClusterInstallConfigApi,
  getClusterInstallPackageApi,
} from 'src/core/api/integration'
import { Button, Typography } from 'antd'
import { IoCloudDownloadOutline } from 'react-icons/io5'
import { Trans, useTranslation } from 'react-i18next'
import 'github-markdown-css/github-markdown.css'
import CopyPre from 'src/core/components/CopyPre'
const decodeBase64 = (base64Str: string) => {
  try {
    const fixedBase64 = base64Str.replace(/-/g, '+').replace(/_/g, '/')
    const binaryStr = atob(fixedBase64)
    const bytes = new Uint8Array(binaryStr.length)
    for (let i = 0; i < binaryStr.length; i++) {
      bytes[i] = binaryStr.charCodeAt(i)
    }
    return new TextDecoder('utf-8').decode(bytes)
  } catch (error) {
    console.error('Base64 error:', error)
  }
}
function getAPOChartVersion() {
  try {
    const config = window.__APP_CONFIG__ || {}
    return config.apoChartVersion || '1.11'
  } catch (e) {
    return '1.11'
  }
}
const InstallCmd = ({ clusterId, clusterType, apoCollector }) => {
  const { t } = useTranslation('core/dataIntegration')
  const [markdownContent, setMarkdownContent] = useState('')
  const downloadFile = (response, suffix = 'yaml') => {
    const contentDisposition = response.headers['content-disposition']
    let filename = 'downloaded-file.yaml'

    if (contentDisposition) {
      const match = contentDisposition.match(/filename="?([^"]+)"?/)
      if (match && match[1]) {
        filename = match[1]
      }
    } else if (clusterType) {
      filename = clusterType + '.' + suffix
    }
    const blob = new Blob([response.data], { type: response.headers['content-type'] })
    const url = window.URL.createObjectURL(blob)

    const a = document.createElement('a')
    a.href = url
    a.download = filename
    document.body.appendChild(a)
    a.click()

    window.URL.revokeObjectURL(url)
    document.body.removeChild(a)
  }
  async function getConfig() {
    try {
      const response = await getClusterInstallConfigApi(clusterId)

      downloadFile(response)
    } catch (error) {
      console.error('error', error)
    }
  }
  const getPackage = async () => {
    try {
      const response = await getClusterInstallPackageApi(clusterId, clusterType)

      downloadFile(response, 'gz')
    } catch (error) {
      console.error('下载失败:', error)
    }
  }

  useEffect(() => {
    getClusterInstallCmdApi(clusterId)
      .then((res) => {
        const decodedMD = decodeBase64(res.installMd)
        setMarkdownContent(decodedMD)
      })
      .catch((error) => {
        setMarkdownContent(t('installCmd.loadError'))
      })
  }, [clusterId])
  const nst = (name: string) => {
    return t(`installCmd.${name}`)
  }
  const getK8sCommand1 = () => {
    return `curl -Lo apo-one-agent-values.yaml http://${apoCollector?.collectorGatewayAddr}:${apoCollector?.ports?.apoBackend || 31363}/api/integration/cluster/install/config?clusterId=${clusterId}`
  }
  const chartVersion = getAPOChartVersion()
  const getK8sCommand2 = () => {
    return `helm repo add apo https://apo-charts.oss-cn-hangzhou.aliyuncs.com
helm repo update apo
helm install apo-one-agent apo/apo-one-agent -n apo --create-namespace --version ${chartVersion} -f apo-one-agent-values.yaml`
  }
  const getVmCommand1 = () => {
    return `curl -Lo installCfg.sh http://${apoCollector?.collectorGatewayAddr}:${apoCollector?.ports?.apoBackend || 31363}/api/integration/cluster/install/config?clusterId=${clusterId}`
  }
  const deployVersion = 'v1.11.000'
  const appVersion = 'v1.11.0'
  const getVmCommand2 = () => {
    return `curl -Lo apo-one-agent-compose-amd64-${deployVersion}.tgz https://apo-ce.oss-cn-hangzhou.aliyuncs.com/apo-one-agent-compose-amd64-${deployVersion}.tgz`
  }
  return (
    <div className="px-3 mx-auto h-full overflow-auto flex flex-col">
      <Typography.Title level={2}>{nst('install')}</Typography.Title>
      {clusterType === 'k8s' ? (
        <Typography>
          <Typography.Title level={3} className="mt-1">
            {nst('helmInstall')}
          </Typography.Title>
          <Typography.Title level={4} className="mt-1">
            {nst('prereq')}
          </Typography.Title>
          <Typography>{nst('prereqDesc')}</Typography>
          <ul className="px-3 pt-2">
            <li>{nst('k8sVer')}</li>
            <li>{nst('helmVer')}</li>
            <li>{nst('admin')}</li>
          </ul>

          <Typography.Title level={4}>{nst('step1')}</Typography.Title>
          <Typography>{nst('downloadWays')}</Typography>
          <ol>
            <li>
              {nst('way1')}
              <CopyPre iconText="" code={getK8sCommand1()} />
            </li>
            <li>
              <Typography>{nst('way2')}</Typography>
              <Button
                type="primary"
                icon={<IoCloudDownloadOutline />}
                onClick={getConfig}
                className="my-2"
              >
                {nst('helmConfigFile')}
              </Button>
            </li>
          </ol>

          <Typography.Title level={4}>{nst('step2')}</Typography.Title>
          <Typography>{nst('executeCmd')}</Typography>
          <CopyPre iconText="" code={getK8sCommand2()} />

          <blockquote>
            <p>
              <strong>{nst('paramDesc')}</strong>
            </p>
            <ul>
              <li>
                <code>-n apo</code>：{nst('paramN')}
                <code>apo</code>。
              </li>
              <li>
                <code>--create-namespace</code>：
                <Trans
                  t={t}
                  i18nKey="installCmd.paramCreateNamespace"
                  components={{
                    // icon: <SiComma />,
                    a: <code>apo</code>,
                  }}
                ></Trans>
              </li>
              <li>
                <code>-f apo-one-agent-values.yaml</code>：{nst('paramF')}
              </li>
            </ul>
          </blockquote>

          <Typography.Title level={4}>{nst('verify')}</Typography.Title>
          <Typography>{nst('verifyCmd')}</Typography>
          <CopyPre iconText="" code={'kubectl get pods -n apo'} />
          <Typography>{nst('successMsg')}</Typography>
        </Typography>
      ) : (
        <Typography>
          <Typography.Title level={3}>{nst('dockerInstall')}</Typography.Title>
          <Typography>{nst('dockerDesc')}</Typography>
          <Typography.Title level={4}>{nst('step1Docker')}</Typography.Title>
          <Typography>{nst('chooseDownload')}</Typography>

          <ul>
            <li>
              {nst('way1')}
              <CopyPre iconText="" code={getVmCommand1()} />
            </li>
            <li>
              <Typography>{nst('way2')}</Typography>
              <Button
                type="primary"
                icon={<IoCloudDownloadOutline />}
                onClick={getConfig}
                className="my-2"
              >
                {nst('btnHelm')}
              </Button>
            </li>
          </ul>

          <Typography.Title level={4}>{nst('step2Docker')}</Typography.Title>
          <Typography>{nst('downloadPackage')}</Typography>
          <CopyPre iconText="" code={getVmCommand2()} />

          <Typography.Title level={4}>{nst('step3Docker')}</Typography.Title>
          <Typography>{nst('runInDir')}</Typography>
          <CopyPre
            iconText=""
            code={'sudo chmod +x ./installCfg.sh && sudo bash ./installCfg.sh'}
          />

          <Typography.Title level={4}>{nst('verify')}</Typography.Title>
          <Typography>{nst('verifyDocker')}</Typography>
          <CopyPre iconText="" code={'docker ps'} />
          <Typography>{nst('successMsgDocker')}</Typography>
        </Typography>
      )}
    </div>
  )
}

export default InstallCmd
