import React, { useEffect, useRef, useState } from 'react';
import { NavLink, useLocation, useNavigate } from 'react-router-dom';
import { useSelector, useDispatch } from 'react-redux';
import logo from 'src/core/assets/brand/logo.svg';
import { CContainer, CHeader, CHeaderNav, useColorModes, CImage } from '@coreui/react';
import { AppBreadcrumb } from './index';
import { AppHeaderDropdown } from './header/index';
import DateTimeRangePicker from './DateTime/DateTimeRangePicker';
import routes from 'src/routes';
import CoachMask from './Mask/CoachMask';
import DateTimeCombine from './DateTime/DateTimeCombine';
import { ConfigProvider, Menu } from 'antd';
import { commercialNav } from 'src/_nav';
import ToolBox from './UserPage/component/ToolBox';
import { HiUserCircle } from "react-icons/hi";
import { RxTriangleDown } from "react-icons/rx";
import { RxTriangleLeft } from "react-icons/rx";
import { Button } from 'antd';
import { getUserInfoApi } from '../api/user';
import { showToast } from '../utils/toast';

const AppHeader = ({ type = 'default' }) => {
  const location = useLocation();
  const navigate = useNavigate();
  const headerRef = useRef();
  const { colorMode, setColorMode } = useColorModes('coreui-free-react-admin-template-theme');
  const [toolVisibal, setToolVisibal] = useState(false);
  const [username, setUsername] = useState("");
  const dispatch = useDispatch();
  const sidebarShow = useSelector((state) => state.sidebarShow);
  const [selectedKeys, setSelectedKeys] = useState([]);
  const toolBoxRef = useRef(null);
  const buttonRef = useRef(null);

  const onClick = ({ item, key, keyPath, domEvent }) => {
    navigate(item.props.to);
  };

  const getItemKey = (navList) => {
    let result = [];
    navList.forEach((item) => {
      if (location.pathname.startsWith(item.to)) {
        result.push(item.key);
      }
      if (item.children) {
        result = result.concat(getItemKey(item.children));
      }
    });
    return result;
  };

  const checkRoute = () => {
    const currentRoute = routes.find((route) => {
      const routePattern = route.path.replace(/:\w+/g, '[^/]+');
      const regex = new RegExp(`^${routePattern}$`);
      return regex.test(location.pathname);
    });
    return !currentRoute?.hideSystemTimeRangePicker;
  };

  useEffect(() => {
    const result = getItemKey(commercialNav);
    setSelectedKeys(result);
  }, [location.pathname]);

  useEffect(() => {
    document.addEventListener('scroll', () => {
      headerRef.current &&
        headerRef.current.classList.toggle('shadow-sm', document.documentElement.scrollTop > 0);
    });
  }, []);

  useEffect(() => {
    const handleClickOutside = (event) => {
      if (
        toolBoxRef.current &&
        !toolBoxRef.current.contains(event.target) &&
        buttonRef.current &&
        !buttonRef.current.contains(event.target)
      ) {
        setToolVisibal(false);
      }
    };

    document.addEventListener('click', handleClickOutside);

    return () => {
      document.removeEventListener('click', handleClickOutside);
    };
  }, []);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          const user = JSON.parse(localStorage.getItem("user"));
          if (user) {
            setUsername(user.username);
          } else {
            setUsername("获取用户信息失败");
          }
        }
      },
      { threshold: 0.1 }
    );

    if (headerRef.current) {
      observer.observe(headerRef.current);
    }

    return () => {
      if (headerRef.current) {
        observer.unobserve(headerRef.current);
      }
    };
  }, []);

  useEffect(() => {
    const handleClickOutside = (event) => {
      if (toolBoxRef.current && !toolBoxRef.current.contains(event.target) &&
        buttonRef.current && !buttonRef.current.contains(event.target)) {
        setToolVisibal(false)
      }
    }

    // 监听点击事件
    document.addEventListener('click', handleClickOutside)

    // 清理事件监听器
    return () => {
      document.removeEventListener('click', handleClickOutside)
    }
  }, [])

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          const user = JSON.parse(localStorage.getItem("user"))
          if (user) {
            setUsername(user.username)
          } else {
            setUsername("获取用户信息失败")
          }
        }
      },
      { threshold: 0.1 }
    );

    if (headerRef.current) {
      observer.observe(headerRef.current);
    }

    return () => {
      if (headerRef.current) {
        observer.unobserve(headerRef.current);
      }
    };
  }, []);

  const vars = {
    borderBottom: 0,
    zIndex: 998,
  };

  return (
    <CHeader position="sticky" className="mb-1 p-0" ref={headerRef} style={vars}>
      <div className="flex justify-between items-center w-full">
        {type === 'united' ? (
          <div className="flex items-center">
            <div className="h-[50px] flex overflow-hidden items-center mr-5">
              <CImage src={logo} className="w-[42px] sidebar-brand-narrow flex-shrink-0 mx-3" />
              <span className="flex-shrink-0 text-lg">向导式可观测平台</span>
            </div>
            <Menu
              mode="horizontal"
              theme="dark"
              items={commercialNav}
              onClick={onClick}
              selectedKeys={selectedKeys}
            ></Menu>
          </div>
        ) : (
          <CHeaderNav className="d-none d-md-flex px-4 py-2 text-base flex items-center h-[50px] flex-grow">
            <AppBreadcrumb />
          </CHeaderNav>
        )}
        <CHeaderNav className="pr-4">
          {location.pathname === '/service/info' && <CoachMask />}
          {checkRoute() && <DateTimeCombine />}
        </CHeaderNav>
        <ConfigProvider
          theme={{
            components: {
              Button: {
                defaultHoverBg: '#30333C',
                defaultBg: '#1E222B',
                defaultBorderColor: 'transparent',
                defaultHoverBorderColor: 'transparent',
                defaultActiveBorderColor: 'transparent',
                defaultActiveBg: '#1E222B',
                defaultHoverColor: 'none',
                defaultShadow: 'none',
                defaultActiveColor: 'none',
              },
            },
          }}
        >
          <div
            className="relative flex items-center select-none w-20 mr-2 rounded-md hover:bg-[#30333C]"
            ref={buttonRef}
            onMouseEnter={() => {
              clearTimeout(buttonRef.current?.hideTimer);
              buttonRef.current.showTimer = setTimeout(() => {
                setToolVisibal(true);
              }, 100); // 延时显示
            }}
            onMouseLeave={() => {
              clearTimeout(buttonRef.current?.showTimer);
              buttonRef.current.hideTimer = setTimeout(() => {
                setToolVisibal(false);
              }, 300); // 延时隐藏
            }}
          >
            <div>
              <HiUserCircle className="w-8 h-8" />
            </div>
            <div className="h-1/2 flex flex-col justify-start">
              <p className="text-base relative -top-0.5">{username}</p>
            </div>
            {toolVisibal ? (
              <RxTriangleDown className="ml-1 mr-4" />
            ) : (
              <RxTriangleLeft className="ml-1 mr-4" />
            )}
            <ToolBox visiable={toolVisibal} setVisiable={setToolVisibal} ref={toolBoxRef} />
          </div>
        </ConfigProvider>
      </div>
    </CHeader>
  );
};

export default AppHeader;
