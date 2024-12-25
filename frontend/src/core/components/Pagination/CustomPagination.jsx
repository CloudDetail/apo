/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useState } from 'react';
import { FiChevronLeft, FiChevronRight } from "react-icons/fi";
import './index.css'
const CustomPagination = React.memo(function CustomPagination({
    pageSize,
    pageIndex,
    total,
    previousPage,
    gotoPage,
    nextPage,
    maxButtons=6
}) {
  const [pageButton, setPageButton] = useState([]);
  const pageCount = Math.ceil(total / pageSize);
  const pageSizeOption = [
    { label: '10 / 页', value: 10 },
    { label: '20 / 页', value: 20 },
    { label: '30 / 页', value: 30 },
    { label: '50 / 页', value: 50 },
    { label: '100 / 页', value: 100 },
  ];

  useEffect(() => {
    let tempButtons = [];
    const halfMaxButtons = Math.floor(maxButtons / 2);

    if (pageCount <= maxButtons) {
      for (let i = 1; i <= pageCount; i++) {
        tempButtons.push(i);
      }
    } else {
      let startPage = Math.max(1, pageIndex - halfMaxButtons);
      let endPage = Math.min(pageCount, pageIndex + halfMaxButtons);

      if (startPage === 1) {
        endPage = Math.min(pageCount, startPage + maxButtons - 1);
      }

      if (endPage === pageCount) {
        startPage = Math.max(1, endPage - maxButtons + 1);
      }

      if (startPage > 1) {
        tempButtons.push(1);
        if (startPage > 2) {
          tempButtons.push({ label: '...', nextPage: startPage - 1 });
        }
      }

      for (let i = startPage; i <= endPage; i++) {
        tempButtons.push(i);
      }

      if (endPage < pageCount) {
        if (endPage < pageCount - 1) {
          tempButtons.push({ label: '...', nextPage: endPage + 1 });
        }
        tempButtons.push(pageCount);
      }
    }

    setPageButton(tempButtons);
  }, [pageCount, pageIndex, maxButtons]);
  return (
    <div className="pagination">
      <div className="basic-pagination">
        <button onClick={previousPage} disabled={pageIndex === 1}>
          <FiChevronLeft />
        </button>
        {pageButton.map((btn, index) => (
          <button
            key={index}
            className={btn === pageIndex ? 'active' : ''}
            onClick={() => {
              if (typeof btn === 'number') {
                gotoPage(btn);
              } else if (btn.label === '...') {
                gotoPage(btn.nextPage);
              }
            }}
          >
            {typeof btn === 'number' ? btn : btn?.label}
          </button>
        ))}
        <button onClick={nextPage} disabled={pageIndex >= pageCount}>
          <FiChevronRight />
        </button>
      </div>
    </div>
  );
});

CustomPagination.displayName = 'CustomPagination';

export default CustomPagination;
