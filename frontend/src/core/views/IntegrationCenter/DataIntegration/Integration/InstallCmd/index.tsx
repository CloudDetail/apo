import { useEffect, useState } from 'react'
import {
  getClusterInstallCmdApi,
  getClusterInstallConfigApi,
  getClusterInstallPackageApi,
} from 'src/core/api/integration'
import ReactMarkdown from 'react-markdown'
import { Button, Card, Typography } from 'antd'
import { IoCloudDownloadOutline } from 'react-icons/io5'
import { useTranslation } from 'react-i18next'

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

const InstallCmd = ({ clusterId, clusterType }) => {
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

  return (
    <div className="p-3 mx-auto markdown-body">
      <Card title={t('installCmd.onlineInstallation')} className="mb-3">
        <div className="p-1">{t('installCmd.downloadHelmConfig')}</div>
        <Button
          type="primary"
          icon={<IoCloudDownloadOutline />}
          onClick={getConfig}
          className="ml-4"
        >
          {t('installCmd.helmConfigFile')}
        </Button>
        <div className="p-1">{t('installCmd.runInstallationCommand')}</div>
        <ReactMarkdown>{markdownContent}</ReactMarkdown>
      </Card>
      <Card title={t('installCmd.offlineInstallation')}>
        <div className="p-1">{t('installCmd.downloadHelmPackage')}</div>
        <Button
          type="primary"
          icon={<IoCloudDownloadOutline />}
          onClick={getPackage}
          className="ml-4"
        >
          {t('installCmd.helmPackageFile')}
        </Button>
        <div className="p-1">{t('installCmd.importOfflineImage')}</div>
        <div className="p-1">{t('installCmd.executeOfflineCommand')}</div>
        <ReactMarkdown>{markdownContent}</ReactMarkdown>
      </Card>
    </div>
  )
}

export default InstallCmd
