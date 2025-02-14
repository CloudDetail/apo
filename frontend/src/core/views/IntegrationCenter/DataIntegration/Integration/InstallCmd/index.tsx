import { useEffect, useState } from 'react'
import {
  getClusterInstallCmdApi,
  getClusterInstallConfigApi,
  getClusterInstallPackageApi,
} from 'src/core/api/integration'
import ReactMarkdown from 'react-markdown'
import { Button } from 'antd'
import { IoCloudDownloadOutline } from 'react-icons/io5'

const decodeBase64 = (base64Str: string) => {
  try {
    // 处理 URL 安全 Base64
    const fixedBase64 = base64Str.replace(/-/g, '+').replace(/_/g, '/')

    // Base64 -> 字符串
    const binaryStr = atob(fixedBase64)

    // 处理 UTF-8
    const bytes = new Uint8Array(binaryStr.length)
    for (let i = 0; i < binaryStr.length; i++) {
      bytes[i] = binaryStr.charCodeAt(i)
    }
    return new TextDecoder('utf-8').decode(bytes)
  } catch (error) {
    console.error('Base64 解码失败:', error)
  }
}

const InstallCmd = ({ clusterId, clusterType }) => {
  const [markdownContent, setMarkdownContent] = useState('')
  const downloadFile = (response, suffix = 'yaml') => {
    // 获取文件名
    const contentDisposition = response.headers['content-disposition']
    let filename = 'downloaded-file.yaml' // 默认文件名

    if (contentDisposition) {
      const match = contentDisposition.match(/filename="?([^"]+)"?/)
      if (match && match[1]) {
        filename = match[1]
      }
    } else if (clusterType) {
      filename = clusterType + '.' + suffix
    }
    // 创建 blob 链接
    const blob = new Blob([response.data], { type: response.headers['content-type'] })
    const url = window.URL.createObjectURL(blob)

    // 创建 <a> 标签，触发下载
    const a = document.createElement('a')
    a.href = url
    a.download = filename // 设置下载的文件名
    document.body.appendChild(a)
    a.click()

    // 释放资源
    window.URL.revokeObjectURL(url)
    document.body.removeChild(a)
  }
  async function getConfig() {
    try {
      const response = await getClusterInstallConfigApi(clusterId)

      downloadFile(response)
    } catch (error) {
      console.error('下载文件失败', error)
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
        console.error('API 请求失败:', error)
        setMarkdownContent('加载失败，请稍后重试！')
      })
  }, [clusterId])

  return (
    <div className="p-6 max-w-2xl mx-auto markdown-body">
      <h2 className="text-xl font-bold mb-4">安装命令</h2>
      <ReactMarkdown>{markdownContent}</ReactMarkdown>

      <div className="flex w-full justify-around my-4">
        <Button type="primary" icon={<IoCloudDownloadOutline />} onClick={getConfig}>
          集群安装配置
        </Button>
        <Button type="primary" icon={<IoCloudDownloadOutline />} onClick={getPackage}>
          集群安装文件
        </Button>
      </div>
    </div>
  )
}

export default InstallCmd
