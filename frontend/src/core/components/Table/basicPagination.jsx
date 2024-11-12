import React, { useEffect, useState } from 'react';
import { FiChevronLeft, FiChevronRight } from "react-icons/fi";
const BasicPagination = React.memo(function BasicPagination({
    pageSize,
    pageIndex,
  page,
  pageCount,
  previousPage,
  gotoPage,
  nextPage,
  setPageSize,
}) {
  const [pageButton, setPageButton] = useState([]);
  const pageSizeOption = [
    {
      label: '10 / 页',
      value: 10,
    },
    {
      label: '20 / 页',
      value: 20,
    },
    {
      label: '30 / 页',
      value: 30,
    },
    {
      label: '50 / 页',
      value: 50,
    },
    {
      label: '100 / 页',
      value: 100,
    },
  ];
  useEffect(() => {
    let tempButtons = [];
    // let pageCount = pagination?.pageCount ? pagination.pageCount : undefined
    // console.log(pageCount)
    if (pageCount <= 9) {
      for (let i = 1; i <= pageCount; i++) {
        tempButtons.push(i);
      }
    } else {
      const startMiddle = pageIndex - 1;
      const endMiddle = pageIndex + 3;

      if (pageIndex <= 3) {
        for (let i = 1; i <= 5; i++) {
          tempButtons.push(i);
        }
        tempButtons.push({ label: '...', nextPage: 6 });
        tempButtons.push(pageCount);
      } else if (pageIndex >= pageCount - 4) {
        tempButtons.push(1);
        tempButtons.push({ label: '...', nextPage: pageCount - 5 });
        for (let i = pageCount - 4; i <= pageCount; i++) {
          tempButtons.push(i);
        }
      } else {
        tempButtons.push(1);
        tempButtons.push({ label: '...', nextPage: pageIndex - 2 });
        for (let i = startMiddle; i <= endMiddle; i++) {
          tempButtons.push(i);
        }
        tempButtons.push({ label: '...', nextPage: pageIndex + 4 });
        tempButtons.push(pageCount);
      }
    }
    // console.log(tempButtons);
    setPageButton(tempButtons);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [pageCount,pageIndex]);
  // 渲染单元格内容
  return (
    <>
        <div className="pagination" >
          <div className="basic-table-pagination">
            <button
              onClick={() => {
                previousPage();
              }}
              disabled={pageIndex === 0}
            >
              <FiChevronLeft />
            </button>
            {pageButton.map((btn, index) => (
              <button
                key={index}
                className={btn === pageIndex + 1 ? 'active' : ''}
                onClick={() => {
                  if (typeof btn === 'number') {
                    gotoPage(btn - 1);
                  } else if (btn.label === '...') {
                    gotoPage(btn.nextPage - 1);
                  }
                }}
              >
                {typeof btn === 'number' ? btn : btn?.label}
              </button>
            ))}
            <button
              onClick={() => {
                nextPage();
              }}
              disabled={pageIndex + 1 >= pageCount}
            >
             <FiChevronRight />
            </button>

          </div>
        </div>
    </>
  );
});


BasicPagination.displayName = 'BasicPagination';

export default BasicPagination;
