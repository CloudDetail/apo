/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import styles from './index.module.scss'
const GlassCard = ({ content }) => {
  return (
    <div className={styles.glassContainer}>
      <div className={styles.glassEffect}>
        <div>
          {/* <h1>毛玻璃效果</h1> */}
          {content}
        </div>
      </div>
    </div>
  )
}
export default GlassCard
