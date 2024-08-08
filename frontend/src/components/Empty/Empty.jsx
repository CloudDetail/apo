import { CImage } from '@coreui/react';
import React from 'react';
import emptyImg from 'src/assets/images/empty.svg'
function Empty(){
    return <div className='w-full h-full flex flex-col justify-center items-center py-5'>
        <CImage src={emptyImg} width={100}/>
        暂无数据
    </div>

}
export default Empty;