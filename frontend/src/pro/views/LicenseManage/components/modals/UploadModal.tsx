import React, { useState } from 'react'
import { Modal, Input, Button, Upload } from 'antd'
import { UploadOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { notify } from 'src/core/utils/notify'
import { uploadLicense } from 'src/pro/api/license'
import { useApiParams } from 'src/core/hooks/useApiParams'
import { useLicense } from 'src/pro/contexts/LicenseContext'

const { TextArea } = Input

const UploadModal = ({ visible, onClose, onUploadSuccess }) => {
  const [licenseText, setLicenseText] = useState('')
  const [loading, setLoading] = useState(false)
  const { sendRequest: fetchRolesRequest } = useApiParams(uploadLicense)
  const { fetchLicenseInfo } = useLicense()
  const { t } = useTranslation('pro/licenseManage')

  const handleUpload = async () => {
    if (!licenseText.trim()) {
      notify({ message: t('upload.warn'), type: 'error' })
      return
    }
    setLoading(true)

    try {
      // 解析 license 文本为对象
      const licenseData = JSON.parse(licenseText)
      console.log(licenseData)
      // 检查必需字段是否存在，并处理字段名映射
      if (
        !licenseData.license ||
        !licenseData.encryptPublicKey ||
        !(licenseData.decryptPrivateKey || licenseData.privateKey) ||
        !licenseData.signPublicKey
      ) {
        notify({ message: t('upload.invalidLicenseFormat'), type: 'error' })
        setLoading(false)
        return
      }

      // 构造符合后端期望的数据结构
      const formattedData = {
        license: licenseData.license,
        encryptPublicKey: licenseData.encryptPublicKey,
        privateKey: licenseData.decryptPrivateKey || licenseData.privateKey, // 使用 decryptPrivateKey 或 privateKey
        signPublicKey: licenseData.signPublicKey,
      }

      // 发送完整的 license 信息
      await fetchRolesRequest(
        { output: JSON.stringify(formattedData) },
        {
          onSuccess: () => {
            notify({ message: t('upload.success'), type: 'success' })
            setLoading(false)
            onUploadSuccess()
          },
          onError: (error) => {
            console.error('Error add role:', error)
            setLoading(false)
            return false
          },
        },
      )
      fetchLicenseInfo()
    } catch (error) {
      notify({ message: t('upload.invalidJson'), type: 'error' })
      setLoading(false)
      return
    }

    // uploadLicense({ licenseText: licenseText })
    //   .then(() => {
    //     notify({ message: t('upload.success'), type: 'success' })
    //     setLoading(false)
    //     onUploadSuccess()
    //   })

    // setTimeout(() => {
    //   notify({ message: t('upload.success'), type: 'success' })
    //   setLoading(false)
    //   onUploadSuccess()
    // }, 1000)
  }

  const beforeUpload = (file: File) => {
    const isLic = file.name.endsWith('.lic')
    if (!isLic) {
      notify({ message: t('upload.error'), type: 'error' })
      return Upload.LIST_IGNORE
    }

    const reader = new FileReader()
    reader.readAsDataURL(file)

    reader.onload = (e) => {
      // get base64 string
      const base64Content = e.target.result

      if (typeof base64Content === 'string') {
        setLicenseText(base64Content) // Safe to call since it's a string
        notify({ message: t('upload.readFileSuccess'), type: 'success' })
      } else {
        notify({ message: t('upload.readFileError'), type: 'error' })
      }
    }
    reader.onerror = () => {
      notify({ message: t('upload.readFileError'), type: 'error' })
    }
    reader.readAsText(file)

    return false
  }

  return (
    <Modal
      title={<p className="text-2xl font-bold">{t('upload.title')}</p>}
      open={visible}
      onCancel={onClose}
      footer={null}
      destroyOnClose
    >
      <div className="flex flex-col space-y-4">
        <p className="my-2 text-base">{t('upload.hint1')}</p>
        <p className="my-2 text-base">
          {t('upload.hint2')}
          <strong>{t('upload.hint3')}</strong>
          {t('upload.hint4')}
        </p>
        <TextArea
          rows={6}
          placeholder={t('upload.placeHolder')}
          value={licenseText}
          onChange={(e) => setLicenseText(e.target.value)}
        />

        <Upload accept=".lic" showUploadList={false} beforeUpload={beforeUpload}>
          <Button icon={<UploadOutlined />}>{t('upload.uploadFile')}</Button>
        </Upload>

        <div className="flex justify-end space-x-4 mt-4">
          <Button onClick={onClose}>{t('upload.cancel')}</Button>
          <Button type="primary" loading={loading} onClick={handleUpload}>
            {t('upload.confirm')}
          </Button>
        </div>
      </div>
    </Modal>
  )
}

export default UploadModal
