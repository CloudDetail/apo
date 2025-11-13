import { Collapse, theme } from 'antd'
import { CSSProperties, useState, useMemo } from 'react'
import { BasicCard } from 'src/core/components/Card/BasicCard'
import { useTranslation } from 'react-i18next'
import LicenseManagementPage from '../LicenseManage'

const text = `
  A dog is a type of domesticated animal.
  Known for its loyalty and faithfulness,
  it can be found: a welcome guest in many households across the world.
`

const PlatformManage = () => {
  const isOp = import.meta.env.VITE_APP_IS_OP === 'true'
  const { t } = useTranslation('pro/platformManage')
  const { token } = theme.useToken()
  const [searchKeyword, setSearchKeyword] = useState('')

  const panelStyle: CSSProperties = useMemo(
    () => ({
      marginTop: 12,
      borderRadius: token.borderRadiusLG,
    }),
    [token.borderRadiusLG],
  )

  const baseItems = [
    // {
    //   key: '1',
    //   label: (
    //     <div className="flex flex-col gap-2 pb-2">
    //       <p className="font-semibold text-2xl flex-nowrap mb-0">
    //         {t("Logo and Title Customization")}
    //       </p>
    //       <p className="font-normal text-base flex flex-nowrap text-[var(--ant-color-text-secondary)] mb-0">
    //         {t("Configure the logo image and system title for the application.")}
    //       </p>
    //     </div>
    //   ),
    //   children: <LogoSetting />,
    //   style: panelStyle,
    // },
  ]

  // const opItems = isOp
  //   ? [
  //       {
  //         key: '2',
  //         label: (
  //           <div className="flex flex-col gap-2 pb-2">
  //             <p className="font-semibold text-2xl flex-nowrap mb-0">{t('License management')}</p>
  //             <p className="font-normal text-base flex flex-nowrap text-[var(--ant-color-text-secondary)] mb-0">
  //               {t('View current system license details and upload new licenses.')}
  //             </p>
  //           </div>
  //         ),
  //         children: <LicenseManagementPage />,
  //         style: panelStyle,
  //       },
  //     ]
  //   : []
  const opItems = [
    {
      key: '2',
      label: (
        <div className="flex flex-col gap-2 pb-2">
          <p className="font-semibold text-2xl flex-nowrap mb-0">{t('License management')}</p>
          <p className="font-normal text-base flex flex-nowrap text-[var(--ant-color-text-secondary)] mb-0">
            {t('View current system license details and upload new licenses.')}
          </p>
        </div>
      ),
      children: <LicenseManagementPage />,
      style: panelStyle,
    },
  ]
  const items = useMemo(() => {
    const allItems = [...baseItems, ...opItems]
    if (!searchKeyword.trim()) return allItems

    const keyword = searchKeyword.toLowerCase()
    return allItems.filter((item) => {
      const labelText =
        `${item.label.props.children[0].props.children} ${item.label.props.children[1].props.children}`.toLowerCase()
      return labelText.includes(keyword)
    })
  }, [searchKeyword, baseItems, opItems])

  return (
    <BasicCard
      bodyStyle={{
        overflow: 'auto',
      }}
    >
      {/* <Input
        placeholder={t('Search page')}
        size="large"
        value={searchKeyword}
        onChange={(e) => setSearchKeyword(e.target.value)}
        style={{
          width: 'calc(100% - 32px)',
          margin: '12px 16px 0 16px',
          borderColor: 'var(--ant-color-border)',
        }}
        prefix={<SearchOutlined className="!text-[var(--ant-color-text-secondary)]" />}
        classNames={{
          prefix: 'text-[var(--ant-color-text-secondary)]',
          input: 'placeholder-[var(--ant-color-text-tertiary)]',
        }}
      /> */}
      <Collapse
        bordered={false}
        defaultActiveKey={['2']}
        style={{ background: token.colorBgContainer }}
        items={items}
        expandIconPosition="end"
        // expandIcon={({ isActive }) => (
        //   <Button>
        //     <span className="text-base font-normal">{isActive ? t('Collapsed') : t('Expand')}</span>
        //   </Button>
        // )}
      />
    </BasicCard>
  )
}

export default PlatformManage
