export const instanceMock = [
  {
    name: 'instance1',
    latency: {
      chartData: {
        1719763200000000: 29777.78226885241,
        1719766800000000: 29621.65289796798,
        1719770400000000: 29558.027823350763,
        1719774000000000: 29507.73824406193,
        1719777600000000: 29474.357013622972,
        1719781200000000: 29416.339069412134,
        1719784800000000: 29355.377021139044,
        1719788400000000: 29366.151845103293,
        1719792000000000: 29260.267739909254,
        1719795600000000: 29205.37895428572,
        1719799200000000: 29080.6442455462,
        1719802800000000: 29093.376491627252,
        1719806400000000: 29092.037945673015,
        1719810000000000: 29078.834009736707,
        1719813600000000: 28959.91698507541,
        1719817200000000: 28952.907105176226,
        1719820800000000: 14491.065358532896,
        1719824400000000: 14506.449258541723,
        1719828000000000: 14498.435534868413,
        1719831600000000: 14434.600187633094,
        1719835200000000: 14408.50042193871,
        1719838800000000: 14361.169800985432,
        1719842400000000: 14361.224331469424,
        1719846000000000: 14337.018328740085,
      },
      value: 36151.010649478034,
      ratio: {
        dayOverDay: 6.03681693223782,
        weekOverDay: null,
      },
    },

    successRate: {
      chartData: {
        1719763200000000: 1,
        1719766800000000: 1,
        1719770400000000: 1,
        1719774000000000: 1,
        1719777600000000: 1,
        1719781200000000: 1,
        1719784800000000: 1,
        1719788400000000: 1,
        1719792000000000: 1,
        1719795600000000: 1,
        1719799200000000: 1,
        1719802800000000: 1,
        1719806400000000: 1,
        1719810000000000: 1,
        1719813600000000: 1,
        1719817200000000: 1,
        1719820800000000: 1,
        1719824400000000: 1,
        1719828000000000: 1,
        1719831600000000: 1,
        1719835200000000: 1,
        1719838800000000: 1,
        1719842400000000: 1,
        1719846000000000: 1,
      },
      value: 1,
      ratio: {
        dayOverDay: 0,
        weekOverDay: null,
      },
    },

    tps: {
      chartData: {
        1719763200000000: 1039098,
        1719766800000000: 1039556,
        1719770400000000: 1039846,
        1719774000000000: 1039900,
        1719777600000000: 1039641,
        1719781200000000: 1039386,
        1719784800000000: 1039451,
        1719788400000000: 1039454,
        1719792000000000: 1039591,
        1719795600000000: 1039347,
        1719799200000000: 1039169,
        1719802800000000: 1038966,
        1719806400000000: 1031053,
        1719810000000000: 1027247,
        1719813600000000: 1027767,
        1719817200000000: 986858,
        1719820800000000: 470696.5,
        1719824400000000: 448036.5,
        1719828000000000: 425385.5,
        1719831600000000: 402914.5,
        1719835200000000: 380387.5,
        1719838800000000: 357813.5,
        1719842400000000: 335138.5,
        1719846000000000: 312597,
      },
      value: 456736,
      ratio: {
        dayOverDay: 376.94412246901203,
        weekOverDay: null,
      },
    },

    logs: {
      chartData: null,
      value: null,
      ratio: {
        dayOverDay: null,
        weekOverDay: null,
      },
    },

    infrastructure: 'success',
    net: {
      status: 'error',
      to: '',
    },
    k8s: {
      status: 'error',
      to: '',
    },
  },
  {
    name: 'instance2',
    latency: {
      chartData: null,
      value: null,
      ratio: {
        dayOverDay: null,
        weekOverDay: null,
      },
    },

    successRate: {
      chartData: null,
      value: null,
      ratio: {
        dayOverDay: null,
        weekOverDay: null,
      },
    },

    tps: {
      chartData: null,
      value: null,
      ratio: {
        dayOverDay: null,
        weekOverDay: null,
      },
    },

    logs: {
      chartData: null,
      value: null,
      ratio: {
        dayOverDay: null,
        weekOverDay: null,
      },
    },

    infrastructure: 'success',
    net: {
      status: 'error',
      to: '',
    },
    k8s: {
      status: 'error',
      to: '',
    },
  },
  {
    name: 'instance3',
    latency: {
      chartData: null,
      value: null,
      ratio: {
        dayOverDay: null,
        weekOverDay: null,
      },
    },

    successRate: {
      chartData: null,
      value: null,
      ratio: {
        dayOverDay: null,
        weekOverDay: null,
      },
    },

    tps: {
      chartData: null,
      value: null,
      ratio: {
        dayOverDay: null,
        weekOverDay: null,
      },
    },

    logs: {
      chartData: null,
      value: null,
      ratio: {
        dayOverDay: null,
        weekOverDay: null,
      },
    },

    infrastructure: 'success',
    net: {
      status: 'error',
      to: '',
    },
    k8s: {
      status: 'error',
      to: '',
    },
  },
]
export const errorInstanceMock = [
  {
    name: 'instance4',
    logs: {
      ratio: {
        dayOverDay: null,
        weekOverDay: null,
      },
    },
    error: 'ts-service HTTP ERROR CODE: 5xx',
    chain: {
      name: 'node1',
      logs: [
        {
          timestamp: '2024-06-20 11:50:39',
          value:
            '[WARN] Install JVM debug symbols to improve profile accuracy 2024-06-20 11:50:39.332 [http-nio-12346-exec-29] INFO  travel.controller.TravelController - [query][Query TripResponse]',
        },
        {
          timestamp: '2024-06-20 11:50:55',
          value:
            '[WARN] Install JVM debug symbols to improve profile accuracy 2024-06-20 11:50:39.332 [http-nio-12346-exec-29] INFO  travel.controller.TravelController - [query][Query TripResponse]',
        },
      ],
      children: [
        {
          name: 'node2',
          logs: [
            {
              timestamp: '2024-06-20 11:50:39',
              value:
                '[WARN] Install JVM debug symbols to improve profile accuracy 2024-06-20 11:50:39.332 [http-nio-12346-exec-29] INFO  travel.controller.TravelController - [query][Query TripResponse]',
            },
            {
              timestamp: '2024-06-20 11:50:55',
              value:
                '[WARN] Install JVM debug symbols to improve profile accuracy 2024-06-20 11:50:39.332 [http-nio-12346-exec-29] INFO  travel.controller.TravelController - [query][Query TripResponse]',
            },
          ],
        },
      ],
    },
  },
  {
    name: 'instance5',
    logs: {
      ratio: {
        dayOverDay: null,
        weekOverDay: null,
      },
    },
    error: 'ts-service HTTP ERROR CODE: 5xx',
    chain: {
      name: 'node2',
      logs: [
        {
          timestamp: '2024-06-20 11:50:39',
          value:
            '[WARN] Install JVM debug symbols to improve profile accuracy 2024-06-20 11:50:39.332 [http-nio-12346-exec-29] INFO  travel.controller.TravelController - [query][Query TripResponse]',
        },
        {
          timestamp: '2024-06-20 11:50:55',
          value:
            '[WARN] Install JVM debug symbols to improve profile accuracy 2024-06-20 11:50:39.332 [http-nio-12346-exec-29] INFO  travel.controller.TravelController - [query][Query TripResponse]',
        },
      ],
      children: [
        {
          name: 'node3',
          logs: [
            {
              timestamp: '2024-06-20 11:50:39',
              value:
                '[WARN] Install JVM debug symbols to improve profile accuracy 2024-06-20 11:50:39.332 [http-nio-12346-exec-29] INFO  travel.controller.TravelController - [query][Query TripResponse]',
            },
            {
              timestamp: '2024-06-20 11:50:55',
              value:
                '[WARN] Install JVM debug symbols to improve profile accuracy 2024-06-20 11:50:39.332 [http-nio-12346-exec-29] INFO  travel.controller.TravelController - [query][Query TripResponse]',
            },
          ],
        },
      ],
    },
  },
  {
    name: 'instance6',
    logs: {
      ratio: {
        dayOverDay: null,
        weekOverDay: null,
      },
    },
    error: 'ts-service HTTP ERROR CODE: 5xx',
    chain: {
      name: 'node2',
      logs: [
        {
          timestamp: '2024-06-20 11:50:39',
          value:
            '[WARN] Install JVM debug symbols to improve profile accuracy 2024-06-20 11:50:39.332 [http-nio-12346-exec-29] INFO  travel.controller.TravelController - [query][Query TripResponse]',
        },
        {
          timestamp: '2024-06-20 11:50:55',
          value:
            '[WARN] Install JVM debug symbols to improve profile accuracy 2024-06-20 11:50:39.332 [http-nio-12346-exec-29] INFO  travel.controller.TravelController - [query][Query TripResponse]',
        },
      ],
      children: [
        {
          name: 'node3',
          logs: [
            {
              timestamp: '2024-06-20 11:50:39',
              value:
                '[WARN] Install JVM debug symbols to improve profile accuracy 2024-06-20 11:50:39.332 [http-nio-12346-exec-29] INFO  travel.controller.TravelController - [query][Query TripResponse]',
            },
            {
              timestamp: '2024-06-20 11:50:55',
              value:
                '[WARN] Install JVM debug symbols to improve profile accuracy 2024-06-20 11:50:39.332 [http-nio-12346-exec-29] INFO  travel.controller.TravelController - [query][Query TripResponse]',
            },
          ],
        },
      ],
    },
  },
]
export const logsInfoMock = [
  {
    name: 'instance 1',
    logs: {
      list: [],
    },
    p90: [],
    average: [],
    logInfo: [
      {
        timestamp: '2024-06-20 11:50:39',
        value:
          '[WARN] Install JVM debug symbols to improve profile accuracy 2024-06-20 11:50:39.332 [http-nio-12346-exec-29] INFO  travel.controller.TravelController - [query][Query TripResponse]',
      },
      {
        timestamp: '2024-06-20 11:50:55',
        value:
          '[WARN] Install JVM debug symbols to improve profile accuracy 2024-06-20 11:50:39.332 [http-nio-12346-exec-29] INFO  travel.controller.TravelController - [query][Query TripResponse]',
      },
    ],
    traceLog: [
      {
        timestamp: '2024-06-20 11:50:39',
        value:
          '[WARN] Install JVM debug symbols to improve profile accuracy 2024-06-20 11:50:39.332 [http-nio-12346-exec-29] INFO  travel.controller.TravelController - [query][Query TripResponse]',
      },
      {
        timestamp: '2024-06-20 11:50:55',
        value:
          '[WARN] Install JVM debug symbols to improve profile accuracy 2024-06-20 11:50:39.332 [http-nio-12346-exec-29] INFO  travel.controller.TravelController - [query][Query TripResponse]',
      },
    ],
  },
  {
    name: 'instance 2',
    logs: {
      list: [],
    },
    p90: [],
    average: [],
    logInfo: [
      {
        timestamp: '2024-06-20 11:50:39',
        value:
          '[WARN] Install JVM debug symbols to improve profile accuracy 2024-06-20 11:50:39.332 [http-nio-12346-exec-29] INFO  travel.controller.TravelController - [query][Query TripResponse]',
      },
      {
        timestamp: '2024-06-20 11:50:55',
        value:
          '[WARN] Install JVM debug symbols to improve profile accuracy 2024-06-20 11:50:39.332 [http-nio-12346-exec-29] INFO  travel.controller.TravelController - [query][Query TripResponse]',
      },
    ],
    traceLog: [
      {
        timestamp: '2024-06-20 11:50:39',
        value:
          '[WARN] Install JVM debug symbols to improve profile accuracy 2024-06-20 11:50:39.332 [http-nio-12346-exec-29] INFO  travel.controller.TravelController - [query][Query TripResponse]',
      },
      {
        timestamp: '2024-06-20 11:50:55',
        value:
          '[WARN] Install JVM debug symbols to improve profile accuracy 2024-06-20 11:50:39.332 [http-nio-12346-exec-29] INFO  travel.controller.TravelController - [query][Query TripResponse]',
      },
    ],
  },
]

export const dependentTableMock = [
  {
    serviceName: 'ts-travel-service',
    serviceDetail: [
      {
        url: 'url1',
        delay: '自身',
        red: 'success',
        logs: 'success',
        infrastructure: 'success',
        net: 'error',
        k8s: 'success',
        timestamp: '2024-6-20 11:18',
      },

      // {
    ],
  },
  {
    serviceName: 'ts-route-service',
    serviceDetail: [
      {
        url: 'url1',
        delay: '依赖',
        red: 'success',
        logs: 'error',
        infrastructure: 'success',
        net: 'error',
        k8s: 'success',
        timestamp: '2024-6-20 11:18',
      },

      // {
    ],
  },
  {
    serviceName: 'ts-station-service',
    serviceDetail: [
      {
        url: 'url1',
        delay: '依赖',
        red: 'success',
        logs: 'success',
        infrastructure: 'success',
        net: 'error',
        k8s: 'success',
        timestamp: '2024-6-20 11:18',
      },
    ],
  },
  {
    serviceName: 'ts-train-service',
    serviceDetail: [
      {
        url: 'url1',
        delay: '依赖',
        red: 'success',
        logs: 'success',
        infrastructure: 'success',
        net: 'error',
        k8s: 'success',
        timestamp: '2024-6-20 11:18',
      },

      // {
    ],
  },
]
export const logsMockList = [
  {
    timestamp: 1722254280000000,
    name: '1',
  },
]
export const logsMock = `2024-07-02 12:00:56.885 [http-nio-14322-exec-13] INFO  travelplan.service.TravelPlanServiceImpl - [getRestTicketNumber][Seat Request][Seat Request is: Seat(travelDate=2024-07-02 00:00:00, trainNumber=G1235, startStation=shanghai, destStation=nanjing, seatType=2, totalNum=2147483647, stations=[nanjing, shanghai])]
2024-07-02 12:00:57.119 [http-nio-14322-exec-13] INFO  travelplan.service.TravelPlanServiceImpl - [getRestTicketNumber][Seat Request][Seat Request is: Seat(travelDate=2024-07-02 00:00:00, trainNumber=G1235, startStation=shanghai, destStation=nanjing, seatType=3, totalNum=2147483647, stations=[nanjing, shanghai])]
2024-07-02 12:00:57.170 [http-nio-14322-exec-6] INFO  travelplan.controller.TravelPlanController - [getCheapest][Search Cheapest][start: nanjing,end: shanghai,time: 2024-07-02 00:00:00]
2024-07-02 12:00:58.279 [http-nio-14322-exec-6] INFO  travelplan.service.TravelPlanServiceImpl - [getRestTicketNumber][Seat Request][Seat Request is: Seat(travelDate=2024-07-02 00:00:00, trainNumber=G1234, startStation=shanghai, destStation=nanjing, seatType=2, totalNum=2147483647, stations=[nanjing, zhenjiang, wuxi, suzhou, shanghai])]
2024-07-02 12:00:58.307 [http-nio-14322-exec-6] INFO  travelplan.service.TravelPlanServiceImpl - [getRestTicketNumber][Seat Request][Seat Request is: Seat(travelDate=2024-07-02 00:00:00, trainNumber=G1234, startStation=shanghai, destStation=nanjing, seatType=3, totalNum=2147483647, stations=[nanjing, zhenjiang, wuxi, suzhou, shanghai])]
2024-07-02 12:00:58.403 [http-nio-14322-exec-16] INFO  travelplan.controller.TravelPlanController - [getCheapest][Search Cheapest][start: nanjing,end: shanghai,time: 2024-07-02 00:00:00]
2024-07-02 12:00:58.721 [http-nio-14322-exec-13] INFO  travelplan.service.TravelPlanServiceImpl - [getRestTicketNumber][Seat Request][Seat Request is: Seat(travelDate=2024-07-02 00:00:00, trainNumber=G1236, startStation=shanghai, destStation=nanjing, seatType=2, totalNum=2147483647, stations=[nanjing, suzhou, shanghai])]
2024-07-02 12:00:58.797 [http-nio-14322-exec-8] INFO  travelplan.controller.TravelPlanController - [getCheapest][Search Cheapest][start: nanjing,end: shanghai,time: 2024-07-02 00:00:00]
2024-07-02 12:00:58.837 [http-nio-14322-exec-8] ERROR o.a.c.c.C.[.[localhost].[/].[dispatcherServlet] - Servlet.service() for servlet [dispatcherServlet] in context with path [] threw exception [Request processing failed; nested exception is org.springframework.web.client.ResourceAccessException: I/O error on POST request for "http://ts-route-plan-service:14578/api/v1/routeplanservice/routePlan/cheapestRoute": ts-route-plan-service; nested exception is java.net.UnknownHostException: ts-route-plan-service] with root cause
java.net.UnknownHostException: ts-route-plan-service
	at java.net.AbstractPlainSocketImpl.connect(AbstractPlainSocketImpl.java:184)
	at java.net.SocksSocketImpl.connect(SocksSocketImpl.java:392)
	at java.net.Socket.connect(Socket.java:589)
	at java.net.Socket.connect(Socket.java:538)
	at sun.net.NetworkClient.doConnect(NetworkClient.java:180)
	at sun.net.www.http.HttpClient.openServer(HttpClient.java:432)
	at sun.net.www.http.HttpClient.openServer(HttpClient.java:527)
	at sun.net.www.http.HttpClient.<init>(HttpClient.java:211)
	at sun.net.www.http.HttpClient.New(HttpClient.java:308)
	at sun.net.www.http.HttpClient.New(HttpClient.java:326)
	at sun.net.www.protocol.http.HttpURLConnection.getNewHttpClient(HttpURLConnection.java:1202)
	at sun.net.www.protocol.http.HttpURLConnection.plainConnect0(HttpURLConnection.java:1138)
	at sun.net.www.protocol.http.HttpURLConnection.plainConnect(HttpURLConnection.java:1032)
	at sun.net.www.protocol.http.HttpURLConnection.connect(HttpURLConnection.java:966)
	at org.springframework.http.client.SimpleBufferingClientHttpRequest.executeInternal(SimpleBufferingClientHttpRequest.java:76)
	at org.springframework.http.client.AbstractBufferingClientHttpRequest.executeInternal(AbstractBufferingClientHttpRequest.java:48)
	at org.springframework.http.client.AbstractClientHttpRequest.execute(AbstractClientHttpRequest.java:53)
	at org.springframework.http.client.InterceptingClientHttpRequest$InterceptingRequestExecution.execute(InterceptingClientHttpRequest.java:109)
	at org.springframework.boot.actuate.metrics.web.client.MetricsClientHttpRequestInterceptor.intercept(MetricsClientHttpRequestInterceptor.java:86)
	at org.springframework.http.client.InterceptingClientHttpRequest$InterceptingRequestExecution.execute(InterceptingClientHttpRequest.java:93)
	at org.springframework.http.client.InterceptingClientHttpRequest.executeInternal(InterceptingClientHttpRequest.java:77)
	at org.springframework.http.client.AbstractBufferingClientHttpRequest.executeInternal(AbstractBufferingClientHttpRequest.java:48)
	at org.springframework.http.client.AbstractClientHttpRequest.execute(AbstractClientHttpRequest.java:53)
	at org.springframework.web.client.RestTemplate.doExecute(RestTemplate.java:737)
	at org.springframework.web.client.RestTemplate.execute(RestTemplate.java:672)
	at org.springframework.web.client.RestTemplate.exchange(RestTemplate.java:610)
	at travelplan.service.TravelPlanServiceImpl.getRoutePlanResultCheapest(TravelPlanServiceImpl.java:255)
	at travelplan.service.TravelPlanServiceImpl.getCheapest(TravelPlanServiceImpl.java:92)
	at travelplan.controller.TravelPlanController.getByCheapest(TravelPlanController.java:41)
	at sun.reflect.GeneratedMethodAccessor59.invoke(Unknown Source)
	at sun.reflect.DelegatingMethodAccessorImpl.invoke(DelegatingMethodAccessorImpl.java:43)
	at java.lang.reflect.Method.invoke(Method.java:498)
	at org.springframework.web.method.support.InvocableHandlerMethod.doInvoke(InvocableHandlerMethod.java:190)
	at org.springframework.web.method.support.InvocableHandlerMethod.invokeForRequest(InvocableHandlerMethod.java:138)
	at org.springframework.web.servlet.mvc.method.annotation.ServletInvocableHandlerMethod.invokeAndHandle(ServletInvocableHandlerMethod.java:105)
	at org.springframework.web.servlet.mvc.method.annotation.RequestMappingHandlerAdapter.invokeHandlerMethod(RequestMappingHandlerAdapter.java:878)
	at org.springframework.web.servlet.mvc.method.annotation.RequestMappingHandlerAdapter.handleInternal(RequestMappingHandlerAdapter.java:792)
	at org.springframework.web.servlet.mvc.method.AbstractHandlerMethodAdapter.handle(AbstractHandlerMethodAdapter.java:87)
	at org.springframework.web.servlet.DispatcherServlet.doDispatch(DispatcherServlet.java:1040)
	at org.springframework.web.servlet.DispatcherServlet.doService(DispatcherServlet.java:943)
	at org.springframework.web.servlet.FrameworkServlet.processRequest(FrameworkServlet.java:1006)
	at org.springframework.web.servlet.FrameworkServlet.doPost(FrameworkServlet.java:909)
	at javax.servlet.http.HttpServlet.service(HttpServlet.java:652)
	at org.springframework.web.servlet.FrameworkServlet.service(FrameworkServlet.java:883)
	at javax.servlet.http.HttpServlet.service(HttpServlet.java:733)
	at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:227)
	at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:162)
	at org.apache.tomcat.websocket.server.WsFilter.doFilter(WsFilter.java:53)
	at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:189)
	at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:162)
	at org.springframework.security.web.FilterChainProxy$VirtualFilterChain.doFilter(FilterChainProxy.java:320)
	at org.springframework.security.web.access.intercept.FilterSecurityInterceptor.invoke(FilterSecurityInterceptor.java:126)
	at org.springframework.security.web.access.intercept.FilterSecurityInterceptor.doFilter(FilterSecurityInterceptor.java:90)
	at org.springframework.security.web.FilterChainProxy$VirtualFilterChain.doFilter(FilterChainProxy.java:334)
	at org.springframework.security.web.access.ExceptionTranslationFilter.doFilter(ExceptionTranslationFilter.java:118)
	at org.springframework.security.web.FilterChainProxy$VirtualFilterChain.doFilter(FilterChainProxy.java:334)
	at org.springframework.security.web.session.SessionManagementFilter.doFilter(SessionManagementFilter.java:137)
	at org.springframework.security.web.FilterChainProxy$VirtualFilterChain.doFilter(FilterChainProxy.java:334)
	at org.springframework.security.web.authentication.AnonymousAuthenticationFilter.doFilter(AnonymousAuthenticationFilter.java:111)
	at org.springframework.security.web.FilterChainProxy$VirtualFilterChain.doFilter(FilterChainProxy.java:334)
	at org.springframework.security.web.servletapi.SecurityContextHolderAwareRequestFilter.doFilter(SecurityContextHolderAwareRequestFilter.java:158)
	at org.springframework.security.web.FilterChainProxy$VirtualFilterChain.doFilter(FilterChainProxy.java:334)
	at org.springframework.security.web.savedrequest.RequestCacheAwareFilter.doFilter(RequestCacheAwareFilter.java:63)
	at org.springframework.security.web.FilterChainProxy$VirtualFilterChain.doFilter(FilterChainProxy.java:334)
	at edu.fudan.common.security.jwt.JWTFilter.doFilterInternal(JWTFilter.java:27)
	at org.springframework.web.filter.OncePerRequestFilter.doFilter(OncePerRequestFilter.java:119)
	at org.springframework.security.web.FilterChainProxy$VirtualFilterChain.doFilter(FilterChainProxy.java:334)
	at org.springframework.security.web.authentication.logout.LogoutFilter.doFilter(LogoutFilter.java:116)
	at org.springframework.security.web.FilterChainProxy$VirtualFilterChain.doFilter(FilterChainProxy.java:334)
	at org.springframework.security.web.header.HeaderWriterFilter.doHeadersAfter(HeaderWriterFilter.java:92)
	at org.springframework.security.web.header.HeaderWriterFilter.doFilterInternal(HeaderWriterFilter.java:77)
	at org.springframework.web.filter.OncePerRequestFilter.doFilter(OncePerRequestFilter.java:119)
	at org.springframework.security.web.FilterChainProxy$VirtualFilterChain.doFilter(FilterChainProxy.java:334)
	at org.springframework.security.web.context.SecurityContextPersistenceFilter.doFilter(SecurityContextPersistenceFilter.java:105)
	at org.springframework.security.web.FilterChainProxy$VirtualFilterChain.doFilter(FilterChainProxy.java:334)
	at org.springframework.security.web.context.request.async.WebAsyncManagerIntegrationFilter.doFilterInternal(WebAsyncManagerIntegrationFilter.java:56)
	at org.springframework.web.filter.OncePerRequestFilter.doFilter(OncePerRequestFilter.java:119)
	at org.springframework.security.web.FilterChainProxy$VirtualFilterChain.doFilter(FilterChainProxy.java:334)
	at org.springframework.security.web.FilterChainProxy.doFilterInternal(FilterChainProxy.java:215)
	at org.springframework.security.web.FilterChainProxy.doFilter(FilterChainProxy.java:178)
	at org.springframework.web.filter.DelegatingFilterProxy.invokeDelegate(DelegatingFilterProxy.java:358)
	at org.springframework.web.filter.DelegatingFilterProxy.doFilter(DelegatingFilterProxy.java:271)
	at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:189)
	at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:162)
	at org.springframework.web.filter.RequestContextFilter.doFilterInternal(RequestContextFilter.java:100)
	at org.springframework.web.filter.OncePerRequestFilter.doFilter(OncePerRequestFilter.java:119)
	at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:189)
	at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:162)
	at org.springframework.web.filter.FormContentFilter.doFilterInternal(FormContentFilter.java:93)
	at org.springframework.web.filter.OncePerRequestFilter.doFilter(OncePerRequestFilter.java:119)
	at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:189)
	at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:162)
	at org.springframework.boot.actuate.metrics.web.servlet.WebMvcMetricsFilter.doFilterInternal(WebMvcMetricsFilter.java:97)
	at org.springframework.web.filter.OncePerRequestFilter.doFilter(OncePerRequestFilter.java:119)
	at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:189)
	at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:162)
	at org.springframework.web.filter.CharacterEncodingFilter.doFilterInternal(CharacterEncodingFilter.java:201)
	at org.springframework.web.filter.OncePerRequestFilter.doFilter(OncePerRequestFilter.java:119)
	at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:189)
	at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:162)
	at org.apache.catalina.core.StandardWrapperValve.invoke(StandardWrapperValve.java:202)
	at org.apache.catalina.core.StandardContextValve.invoke(StandardContextValve.java:97)
	at org.apache.catalina.authenticator.AuthenticatorBase.invoke(AuthenticatorBase.java:542)
	at org.apache.catalina.core.StandardHostValve.invoke(StandardHostValve.java:143)
	at org.apache.catalina.valves.ErrorReportValve.invoke(ErrorReportValve.java:92)
	at org.apache.catalina.core.StandardEngineValve.invoke(StandardEngineValve.java:78)
	at org.apache.catalina.valves.RemoteIpValve.invoke(RemoteIpValve.java:764)
	at org.apache.catalina.connector.CoyoteAdapter.service(CoyoteAdapter.java:357)
	at org.apache.coyote.http11.Http11Processor.service(Http11Processor.java:374)
	at org.apache.coyote.AbstractProcessorLight.process(AbstractProcessorLight.java:65)
	at org.apache.coyote.AbstractProtocol$ConnectionHandler.process(AbstractProtocol.java:893)
	at org.apache.tomcat.util.net.NioEndpoint$SocketProcessor.doRun(NioEndpoint.java:1707)
	at org.apache.tomcat.util.net.SocketProcessorBase.run(SocketProcessorBase.java:49)
	at java.util.concurrent.ThreadPoolExecutor.runWorker(ThreadPoolExecutor.java:1142)
	at java.util.concurrent.ThreadPoolExecutor$Worker.run(ThreadPoolExecutor.java:617)
	at org.apache.tomcat.util.threads.TaskThread$WrappingRunnable.run(TaskThread.java:61)
	at java.lang.Thread.run(Thread.java:745)
2024-07-02 12:00:58.891 [http-nio-14322-exec-13] INFO  travelplan.service.TravelPlanServiceImpl - [getRestTicketNumber][Seat Request][Seat Request is: Seat(travelDate=2024-07-02 00:00:00, trainNumber=G1236, startStation=shanghai, destStation=nanjing, seatType=3, totalNum=2147483647, stations=[nanjing, suzhou, shanghai])]
2024-07-02 12:00:59.158 [http-nio-14322-exec-2] INFO  travelplan.controller.TravelPlanController - [getCheapest][Search Cheapest][start: nanjing,end: shanghai,time: 2024-07-02 00:00:00]
2024-07-02 12:00:59.161 [http-nio-14322-exec-2] ERROR o.a.c.c.C.[.[localhost].[/].[dispatcherServlet] - Servlet.service() for servlet [dispatcherServlet] in context with path [] threw exception [Request processing failed; nested exception is org.springframework.web.client.ResourceAccessException: I/O error on POST request for "http://ts-route-plan-service:14578/api/v1/routeplanservice/routePlan/cheapestRoute": ts-route-plan-service; nested exception is java.net.UnknownHostException: ts-route-plan-service] with root cause
java.net.UnknownHostException: ts-route-plan-service
	at java.net.AbstractPlainSocketImpl.connect(AbstractPlainSocketImpl.java:184)
	at java.net.SocksSocketImpl.connect(SocksSocketImpl.java:392)
	at java.net.Socket.connect(Socket.java:589)
	at java.net.Socket.connect(Socket.java:538)
	at sun.net.NetworkClient.doConnect(NetworkClient.java:180)
	at sun.net.www.http.HttpClient.openServer(HttpClient.java:432)
	at sun.net.www.http.HttpClient.openServer(HttpClient.java:527)
	at sun.net.www.http.HttpClient.<init>(HttpClient.java:211)
	at sun.net.www.http.HttpClient.New(HttpClient.java:308)
	at sun.net.www.http.HttpClient.New(HttpClient.java:326)
	at sun.net.www.protocol.http.HttpURLConnection.getNewHttpClient(HttpURLConnection.java:1202)
	at sun.net.www.protocol.http.HttpURLConnection.plainConnect0(HttpURLConnection.java:1138)
	at sun.net.www.protocol.http.HttpURLConnection.plainConnect(HttpURLConnection.java:1032)
	at sun.net.www.protocol.http.HttpURLConnection.connect(HttpURLConnection.java:966)
	at org.springframework.http.client.SimpleBufferingClientHttpRequest.executeInternal(SimpleBufferingClientHttpRequest.java:76)
	at org.springframework.http.client.AbstractBufferingClientHttpRequest.executeInternal(AbstractBufferingClientHttpRequest.java:48)
	at org.springframework.http.client.AbstractClientHttpRequest.execute(AbstractClientHttpRequest.java:53)
	at org.springframework.http.client.InterceptingClientHttpRequest$InterceptingRequestExecution.execute(InterceptingClientHttpRequest.java:109)
	at org.springframework.boot.actuate.metrics.web.client.MetricsClientHttpRequestInterceptor.intercept(MetricsClientHttpRequestInterceptor.java:86)
	at org.springframework.http.client.InterceptingClientHttpRequest$InterceptingRequestExecution.execute(InterceptingClientHttpRequest.java:93)
	at org.springframework.http.client.InterceptingClientHttpRequest.executeInternal(InterceptingClientHttpRequest.java:77)
	at org.springframework.http.client.AbstractBufferingClientHttpRequest.executeInternal(AbstractBufferingClientHttpRequest.java:48)
	at org.springframework.http.client.AbstractClientHttpRequest.execute(AbstractClientHttpRequest.java:53)
	at org.springframework.web.client.RestTemplate.doExecute(RestTemplate.java:737)
	at org.springframework.web.client.RestTemplate.execute(RestTemplate.java:672)
	at org.springframework.web.client.RestTemplate.exchange(RestTemplate.java:610)
	at travelplan.service.TravelPlanServiceImpl.getRoutePlanResultCheapest(TravelPlanServiceImpl.java:255)
	at travelplan.service.TravelPlanServiceImpl.getCheapest(TravelPlanServiceImpl.java:92)
	at travelplan.controller.TravelPlanController.getByCheapest(TravelPlanController.java:41)
	at sun.reflect.GeneratedMethodAccessor59.invoke(Unknown Source)
	at sun.reflect.DelegatingMethodAccessorImpl.invoke(DelegatingMethodAccessorImpl.java:43)
	at java.lang.reflect.Method.invoke(Method.java:498)
	at org.springframework.web.method.support.InvocableHandlerMethod.doInvoke(InvocableHandlerMethod.java:190)
	at org.springframework.web.method.support.InvocableHandlerMethod.invokeForRequest(InvocableHandlerMethod.java:138)
	at org.springframework.web.servlet.mvc.method.annotation.ServletInvocableHandlerMethod.invokeAndHandle(ServletInvocableHandlerMethod.java:105)
	at org.springframework.web.servlet.mvc.method.annotation.RequestMappingHandlerAdapter.invokeHandlerMethod(RequestMappingHandlerAdapter.java:878)
	at org.springframework.web.servlet.mvc.method.annotation.RequestMappingHandlerAdapter.handleInternal(RequestMappingHandlerAdapter.java:792)
	at org.springframework.web.servlet.mvc.method.AbstractHandlerMethodAdapter.handle(AbstractHandlerMethodAdapter.java:87)
	at org.springframework.web.servlet.DispatcherServlet.doDispatch(DispatcherServlet.java:1040)
	at org.springframework.web.servlet.DispatcherServlet.doService(DispatcherServlet.java:943)
	at org.springframework.web.servlet.FrameworkServlet.processRequest(FrameworkServlet.java:1006)
	at org.springframework.web.servlet.FrameworkServlet.doPost(FrameworkServlet.java:909)
	at javax.servlet.http.HttpServlet.service(HttpServlet.java:652)
	at org.springframework.web.servlet.FrameworkServlet.service(FrameworkServlet.java:883)
	at javax.servlet.http.HttpServlet.service(HttpServlet.java:733)
	at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:227)
	at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:162)
	at org.apache.tomcat.websocket.server.WsFilter.doFilter(WsFilter.java:53)
	at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:189)
	at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:162)
	at org.springframework.security.web.FilterChainProxy$VirtualFilterChain.doFilter(FilterChainProxy.java:320)
	at org.springframework.security.web.access.intercept.FilterSecurityInterceptor.invoke(FilterSecurityInterceptor.java:126)
	at org.springframework.security.web.access.intercept.FilterSecurityInterceptor.doFilter(FilterSecurityInterceptor.java:90)
	at org.springframework.security.web.FilterChainProxy$VirtualFilterChain.doFilter(FilterChainProxy.java:334)
	at org.springframework.security.web.access.ExceptionTranslationFilter.doFilter(ExceptionTranslationFilter.java:118)
	at org.springframework.security.web.FilterChainProxy$VirtualFilterChain.doFilter(FilterChainProxy.java:334)
	at org.springframework.security.web.session.SessionManagementFilter.doFilter(SessionManagementFilter.java:137)
	at org.springframework.security.web.FilterChainProxy$VirtualFilterChain.doFilter(FilterChainProxy.java:334)
	at org.springframework.security.web.authentication.AnonymousAuthenticationFilter.doFilter(AnonymousAuthenticationFilter.java:111)
	at org.springframework.security.web.FilterChainProxy$VirtualFilterChain.doFilter(FilterChainProxy.java:334)
	at org.springframework.security.web.servletapi.SecurityContextHolderAwareRequestFilter.doFilter(SecurityContextHolderAwareRequestFilter.java:158)
	at org.springframework.security.web.FilterChainProxy$VirtualFilterChain.doFilter(FilterChainProxy.java:334)
	at org.springframework.security.web.savedrequest.RequestCacheAwareFilter.doFilter(RequestCacheAwareFilter.java:63)
	at org.springframework.security.web.FilterChainProxy$VirtualFilterChain.doFilter(FilterChainProxy.java:334)
	at edu.fudan.common.security.jwt.JWTFilter.doFilterInternal(JWTFilter.java:27)
	at org.springframework.web.filter.OncePerRequestFilter.doFilter(OncePerRequestFilter.java:119)
	at org.springframework.security.web.FilterChainProxy$VirtualFilterChain.doFilter(FilterChainProxy.java:334)
	at org.springframework.security.web.authentication.logout.LogoutFilter.doFilter(LogoutFilter.java:116)
	at org.springframework.security.web.FilterChainProxy$VirtualFilterChain.doFilter(FilterChainProxy.java:334)
	at org.springframework.security.web.header.HeaderWriterFilter.doHeadersAfter(HeaderWriterFilter.java:92)
	at org.springframework.security.web.header.HeaderWriterFilter.doFilterInternal(HeaderWriterFilter.java:77)
	at org.springframework.web.filter.OncePerRequestFilter.doFilter(OncePerRequestFilter.java:119)
	at org.springframework.security.web.FilterChainProxy$VirtualFilterChain.doFilter(FilterChainProxy.java:334)
	at org.springframework.security.web.context.SecurityContextPersistenceFilter.doFilter(SecurityContextPersistenceFilter.java:105)
	at org.springframework.security.web.FilterChainProxy$VirtualFilterChain.doFilter(FilterChainProxy.java:334)
	at org.springframework.security.web.context.request.async.WebAsyncManagerIntegrationFilter.doFilterInternal(WebAsyncManagerIntegrationFilter.java:56)
	at org.springframework.web.filter.OncePerRequestFilter.doFilter(OncePerRequestFilter.java:119)
	at org.springframework.security.web.FilterChainProxy$VirtualFilterChain.doFilter(FilterChainProxy.java:334)
	at org.springframework.security.web.FilterChainProxy.doFilterInternal(FilterChainProxy.java:215)
	at org.springframework.security.web.FilterChainProxy.doFilter(FilterChainProxy.java:178)
	at org.springframework.web.filter.DelegatingFilterProxy.invokeDelegate(DelegatingFilterProxy.java:358)
	at org.springframework.web.filter.DelegatingFilterProxy.doFilter(DelegatingFilterProxy.java:271)
	at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:189)
	at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:162)
	at org.springframework.web.filter.RequestContextFilter.doFilterInternal(RequestContextFilter.java:100)
	at org.springframework.web.filter.OncePerRequestFilter.doFilter(OncePerRequestFilter.java:119)
	at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:189)
	at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:162)
	at org.springframework.web.filter.FormContentFilter.doFilterInternal(FormContentFilter.java:93)
	at org.springframework.web.filter.OncePerRequestFilter.doFilter(OncePerRequestFilter.java:119)
	at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:189)
	at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:162)
	at org.springframework.boot.actuate.metrics.web.servlet.WebMvcMetricsFilter.doFilterInternal(WebMvcMetricsFilter.java:97)
	at org.springframework.web.filter.OncePerRequestFilter.doFilter(OncePerRequestFilter.java:119)
	at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:189)
	at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:162)
	at org.springframework.web.filter.CharacterEncodingFilter.doFilterInternal(CharacterEncodingFilter.java:201)
	at org.springframework.web.filter.OncePerRequestFilter.doFilter(OncePerRequestFilter.java:119)
	at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:189)
	at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:162)
	at org.apache.catalina.core.StandardWrapperValve.invoke(StandardWrapperValve.java:202)
	at org.apache.catalina.core.StandardContextValve.invoke(StandardContextValve.java:97)
	at org.apache.catalina.authenticator.AuthenticatorBase.invoke(AuthenticatorBase.java:542)
	at org.apache.catalina.core.StandardHostValve.invoke(StandardHostValve.java:143)
	at org.apache.catalina.valves.ErrorReportValve.invoke(ErrorReportValve.java:92)
	at org.apache.catalina.core.StandardEngineValve.invoke(StandardEngineValve.java:78)
	at org.apache.catalina.valves.RemoteIpValve.invoke(RemoteIpValve.java:764)
	at org.apache.catalina.connector.CoyoteAdapter.service(CoyoteAdapter.java:357)
	at org.apache.coyote.http11.Http11Processor.service(Http11Processor.java:374)
	at org.apache.coyote.AbstractProcessorLight.process(AbstractProcessorLight.java:65)
	at org.apache.coyote.AbstractProtocol$ConnectionHandler.process(AbstractProtocol.java:893)
	at org.apache.tomcat.util.net.NioEndpoint$SocketProcessor.doRun(NioEndpoint.java:1707)
	at org.apache.tomcat.util.net.SocketProcessorBase.run(SocketProcessorBase.java:49)
	at java.util.concurrent.ThreadPoolExecutor.runWorker(ThreadPoolExecutor.java:1142)
	at java.util.concurrent.ThreadPoolExecutor$Worker.run(ThreadPoolExecutor.java:617)
	at org.apache.tomcat.util.threads.TaskThread$WrappingRunnable.run(TaskThread.java:61)
	at java.lang.Thread.run(Thread.java:745)
2024-07-02 12:00:59.286 [http-nio-14322-exec-16] INFO  travelplan.service.TravelPlanServiceImpl - [getRestTicketNumber][Seat Request][Seat Request is: Seat(travelDate=2024-07-02 00:00:00, trainNumber=G1234, startStation=shanghai, destStation=nanjing, seatType=2, totalNum=2147483647, stations=[nanjing, zhenjiang, wuxi, suzhou, shanghai])]
2024-07-02 12:00:59.325 [http-nio-14322-exec-16] INFO  travelplan.service.TravelPlanServiceImpl - [getRestTicketNumber][Seat Request][Seat Request is: Seat(travelDate=2024-07-02 00:00:00, trainNumber=G1234, startStation=shanghai, destStation=nanjing, seatType=3, totalNum=2147483647, stations=[nanjing, zhenjiang, wuxi, suzhou, shanghai])]
2024-07-02 12:00:59.586 [http-nio-14322-exec-16] INFO  travelplan.service.TravelPlanServiceImpl - [getRestTicketNumber][Seat Request][Seat Request is: Seat(travelDate=2024-07-02 00:00:00, trainNumber=Z1236, startStation=shanghai, destStation=nanjing, seatType=2, totalNum=2147483647, stations=[taiyuan, shijiazhuang, nanjing, shanghai])]
2024-07-02 12:00:59.604 [http-nio-14322-exec-16] INFO  travelplan.service.TravelPlanServiceImpl - [getRestTicketNumber][Seat Request][Seat Request is: Seat(travelDate=2024-07-02 00:00:00, trainNumber=Z1236, startStation=shanghai, destStation=nanjing, seatType=3, totalNum=2147483647, stations=[taiyuan, shijiazhuang, nanjing, shanghai])]
2024-07-02 12:00:59.628 [http-nio-14322-exec-16] INFO  travelplan.service.TravelPlanServiceImpl - [getRestTicketNumber][Seat Request][Seat Request is: Seat(travelDate=2024-07-02 00:00:00, trainNumber=G1235, startStation=shanghai, destStation=nanjing, seatType=2, totalNum=2147483647, stations=[nanjing, shanghai])]
2024-07-02 12:00:59.657 [http-nio-14322-exec-16] INFO  travelplan.service.TravelPlanServiceImpl - [getRestTicketNumber][Seat Request][Seat Request is: Seat(travelDate=2024-07-02 00:00:00, trainNumber=G1235, startStation=shanghai, destStation=nanjing, seatType=3, totalNum=2147483647, stations=[nanjing, shanghai])]
2024-07-02 12:00:59.720 [http-nio-14322-exec-16] INFO  travelplan.service.TravelPlanServiceImpl - [getRestTicketNumber][Seat Request][Seat Request is: Seat(travelDate=2024-07-02 00:00:00, trainNumber=G1236, startStation=shanghai, destStation=nanjing, seatType=2, totalNum=2147483647, stations=[nanjing, suzhou, shanghai])]
2024-07-02 12:00:59.751 [http-nio-14322-exec-16] INFO  travelplan.service.TravelPlanServiceImpl - [getRestTicketNumber][Seat Request][Seat Request is: Seat(travelDate=2024-07-02 00:00:00, trainNumber=G1236, startStation=shanghai, destStation=nanjing, seatType=3, totalNum=2147483647, stations=[nanjing, suzhou, shanghai])]
2024-07-02 16:21:00.039 [http-nio-14322-exec-6] INFO  travelplan.service.TravelPlanServiceImpl - [getRestTicketNumber][Seat Request][Seat Request is: Seat(travelDate=2024-07-02 00:00:00, trainNumber=Z1236, startStation=shanghai, destStation=nanjing, seatType=2, totalNum=2147483647, stations=[taiyuan, shijiazhuang, nanjing, shanghai])]
2024-07-02 16:21:01.346 [http-nio-14322-exec-21] INFO  travelplan.controller.TravelPlanController - [getCheapest][Search Cheapest][start: nanjing,end: shanghai,time: 2024-07-02 00:00:00]`

const chainData = {
  timestamp: 1721988206508001,
  traceId: '66eb67dc3c590c64cd70433a83aa5a98',
  errors: [
    {
      type: 'java.net.ConnectException',
      message: 'Connection refused (Connection refused)',
    },
    {
      type: 'org.springframework.web.client.ResourceAccessException',
      message:
        'I/O error on POST request for "http://ts-basic-service:15680/api/v1/basicservice/basic/travels": Connection refused (Connection refused); nested exception is java.net.ConnectException: Connection refused (Connection refused)',
    },
    {
      type: 'java.net.ConnectException',
      message: 'Connection refused (Connection refused)',
    },
  ],
  parents: [
    {
      instance: 'ts-route-plan-service-598bd5fc56-vflc9',
      isTraced: true,
    },
    {
      instance: 'ts-route-plan-service-598bd5fc56-vflc9112',
      isTraced: true,
    },
  ],
  current: {
    instance: 'ts-travel-service-8576ffcfd5-9879m',
    isTraced: true,
  },
  children: [
    {
      instance: 'ts-route-plan-service-598bd5fc56-vflc911',
      isTraced: true,
    },
    {
      instance: 'ts-route-plan-service-598bd5fc56-vflc922',
      isTraced: true,
    },
    {
      instance: 'ts-route-plan-service-598bd5fc56-vflc9333',
      isTraced: true,
    },
  ],
}
