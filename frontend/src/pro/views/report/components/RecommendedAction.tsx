import { useState } from 'react'
import { SuggestLevel } from '../type'
import { useSelector } from 'react-redux'
import { useTranslation } from 'react-i18next'

interface SuggestItem {
  action: string
  level: SuggestLevel
}

const RecommendedAction = ({ suggest }: { suggest: SuggestItem[] }) => {
  const [checkedActions, setCheckedActions] = useState<Record<number, boolean>>({})
  const toggleAction = (index: number) => {
    setCheckedActions((prev) => ({
      ...prev,
      [index]: !prev[index],
    }))
  }

  const { theme } = useSelector((state) => state.settingReducer)
  const isDark = theme === 'dark'
  const { t } = useTranslation('pro/report')
  return (
    <div className="mt-6 space-y-4">
      {suggest.length === 0 ? (
        <div
          className={`border border-dashed rounded-md p-4 text-center ${
            isDark ? 'border-slate-600 text-slate-400' : 'border-slate-300 text-slate-500'
          }`}
        >
          {t('noSuggest')}
        </div>
      ) : (
        suggest.map((action, index) => {
          const isChecked = checkedActions[index]

          return (
            <div
              key={index}
              className={`border rounded-md px-3 py-2 transition-colors ${
                isChecked
                  ? isDark
                    ? 'border-emerald-800 bg-emerald-900/30'
                    : 'border-emerald-200 bg-emerald-50'
                  : isDark
                    ? 'border-slate-700 hover:border-slate-600'
                    : 'border-slate-200 hover:border-slate-300'
              }`}
            >
              <div className="flex items-start">
                {/* <div
                  className="flex-shrink-0 mr-3 mt-1 cursor-pointer"
                  onClick={() => toggleAction(index)}
                >
                  <div
                    className={`w-5 h-5 rounded border flex items-center justify-center ${
                      isChecked
                        ? isDark
                          ? 'bg-emerald-600 border-emerald-600'
                          : 'bg-emerald-500 border-emerald-500'
                        : isDark
                          ? 'border-slate-600'
                          : 'border-slate-300'
                    }`}
                  >
                    {isChecked && (
                      <svg className="w-3 h-3 text-white" fill="currentColor" viewBox="0 0 20 20">
                        <path
                          fillRule="evenodd"
                          d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                          clipRule="evenodd"
                        />
                      </svg>
                    )}
                  </div>
                </div> */}

                {/* 等级标签 */}
                {action.level && (
                  <span
                    className={`inline-block text-xs mr-1 px-2 py-1 rounded-full ${
                      action.level === 'high'
                        ? isDark
                          ? 'bg-rose-900/50 text-rose-400'
                          : 'bg-rose-100 text-rose-700'
                        : action.level === 'medium'
                          ? isDark
                            ? 'bg-amber-900/50 text-amber-400'
                            : 'bg-amber-100 text-amber-700'
                          : isDark
                            ? 'bg-slate-700 text-slate-300'
                            : 'bg-slate-100 text-slate-700'
                    }`}
                  >
                    {action.level.toUpperCase()}
                  </span>
                )}

                {/* 动作标题 */}
                <div
                  className={`font-medium text-sm ${
                    isChecked
                      ? isDark
                        ? 'text-emerald-400 line-through'
                        : 'text-emerald-700 line-through'
                      : isDark
                        ? 'text-slate-200'
                        : 'text-slate-800'
                  }`}
                >
                  {action.action || `{{ACTION_${index + 1}_TITLE}}`}
                </div>
              </div>
            </div>
          )
        })
      )}
    </div>
  )
}

export default RecommendedAction
