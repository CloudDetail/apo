package model

import (
	"encoding/json"
	"fmt"
	"github.com/magiconair/properties/assert"
	"testing"
)

type Flame struct {
	Version     int         `json:"version"`
	Flamebearer FlameBearer `json:"flamebearer"`
	Metadata    Metadata    `json:"metadata"`
}

type Metadata struct {
	Name       string `json:"name"`
	Format     string `json:"format"`
	SampleRate int    `json:"sampleRate"`
	SpyName    string `json:"spyName"`
	Units      string `json:"units"`
}

func TestToTree(t *testing.T) {
	tree := newTree1()
	flame := NewFlameGraph(tree)
	old := newTree1()
	oldFlame := NewFlameGraph(old)
	mergeTree := &Tree{}
	mergeTree.MergeFlameGraph(flame)
	output := NewFlameGraph(mergeTree)
	assert.Equal(t, output, oldFlame)
}

// TestFlameGroup
// 1. 测试两个节点的子节点不同（1->2 3->4 | 1->4 2->3）√
// 2. 测试偏移相同的节点是否会被覆盖 (1->2 offset30 | 1->3 offset30) √
// 3. 测试相同节点合并（1->2 | 1->2） √
// 4. 测试同级节点顺序不同（1->2,3 | 1->3,2）√
// 5. 测试层级数量不同（1->2 | 1->2->3）√
// 6. 测试两个完全不同的火焰图（1->2 | 3->4）√
func TestFlameGroup(t *testing.T) {
	tree1 := newTree1()
	tree2 := newTree2()
	graph1 := NewFlameGraph(tree1)
	graph2 := NewFlameGraph(tree2)
	fmt.Println(graph1.Names)
	fmt.Println(graph1.Levels)
	fmt.Println(graph2.Names)
	fmt.Println(graph2.Levels)
	mergedTree := &Tree{}
	mergedTree.MergeFlameGraph(graph1)
	mergedTree.MergeFlameGraph(graph2)
	mergedFlame := NewFlameGraph(mergedTree)
	fmt.Println(mergedFlame.Names)
	fmt.Println(mergedFlame.Levels)
	fmt.Println(mergedFlame.NumTicks)
	fmt.Println(mergedFlame.MaxSelf)

	fmt.Println(wrapToFlame(*mergedFlame))
}

func wrapToFlame(flame FlameBearer) string {
	f := Flame{
		Version:     1,
		Flamebearer: flame,
		Metadata:    Metadata{Format: "single"},
	}
	fStr, _ := json.Marshal(f)
	return string(fStr)
}

func newNode(name string, self, total int64) *node {
	return &node{
		name:  name,
		self:  self,
		total: total,
	}
}

func newTree1() *Tree {
	root := newNode("root", 0, 20)

	func1 := newNode("func1", 10, 20)
	func2 := newNode("func2", 10, 10)

	root.children = append(root.children, func1)

	func1.children = append(func1.children, func2)
	tree := &Tree{
		root: []*node{root},
	}

	return tree
}

func newTree2() *Tree {
	root := newNode("root", 0, 70)

	func1 := newNode("func1", 20, 30)
	func2 := newNode("func2", 10, 10)
	func3 := newNode("func3", 30, 40)
	func4 := newNode("func4", 10, 10)
	root.children = append(root.children, func1, func3)

	func1.children = append(func1.children, func2)
	func3.children = append(func3.children, func4)
	tree := &Tree{
		root: []*node{root},
	}

	return tree
}

func TestWrapToFlame(t *testing.T) {
	data := "{\"names\":[\"total\",\"org/springframework/boot/web/embedded/tomcat/TomcatWebServer$1.run\",\"org/apache/catalina/core/StandardServer.await\",\"java/lang/Thread.sleep\",\"libpthread-2.31.so.__pthread_cond_timedwait\",\"org/apache/tomcat/util/net/NioBlockingSelector$BlockPoller.run\",\"sun/nio/ch/SelectorImpl.select\",\"sun/nio/ch/SelectorImpl.lockAndDoSelect\",\"sun/nio/ch/EPollSelectorImpl.doSelect\",\"sun/nio/ch/EPoll.wait\",\"libc-2.31.so.epoll_wait\",\"okio/AsyncTimeout$Watchdog.run\",\"okio/AsyncTimeout$Companion.awaitTimeout\",\"java/util/concurrent/locks/AbstractQueuedSynchronizer$ConditionObject.await\",\"java/util/concurrent/locks/LockSupport.parkNanos\",\"jdk/internal/misc/Unsafe.park\",\"libpthread-2.31.so.start_thread\",\"libjvm.so.thread_native_entry\",\"libjvm.so.Thread::call_run\",\"libjvm.so.WatcherThread::run\",\"libjvm.so.WatcherThread::sleep\",\"libjvm.so.Monitor::wait\",\"libjvm.so.Monitor::IWait\",\"libpthread-2.31.so.pthread_mutex_unlock\",\"libpthread-2.31.so.__pthread_mutex_unlock_usercnt\",\"libjvm.so.JavaThread::thread_main_inner\",\"libjvm.so.attach_listener_thread_entry\",\"libjvm.so.AttachListener::dequeue\",\"libpthread-2.31.so.__libc_accept\",\"libjvm.so.CompileBroker::compiler_thread_loop\",\"libjvm.so.CompileBroker::invoke_compiler_on_method\",\"libjvm.so.ciEnv::ciEnv\",\"libjvm.so.ciObjectFactory::ciObjectFactory\",\"libjvm.so.Compiler::compile_method\",\"libjvm.so.Compilation::Compilation\",\"libjvm.so.Compilation::compile_method\",\"libjvm.so.Compilation::compile_java_method\",\"libjvm.so.Compilation::emit_lir\",\"libjvm.so.LinearScan::do_linear_scan\",\"libjvm.so.LinearScan::allocate_registers\",\"libjvm.so.IntervalWalker::walk_to\",\"libjvm.so.LinearScanWalker::activate_current\",\"libjvm.so.LinearScanWalker::alloc_free_reg\",\"libjvm.so.Compilation::build_hir\",\"libjvm.so.IR::IR\",\"libjvm.so.GraphBuilder::GraphBuilder\",\"libjvm.so.GraphBuilder::state_at_entry\",\"libjvm.so.GraphBuilder::iterate_all_blocks\",\"libjvm.so.GraphBuilder::iterate_bytecodes_for_block\",\"libjvm.so.GraphBuilder::invoke\",\"libjvm.so.GraphBuilder::try_inline\",\"libjvm.so.GraphBuilder::try_inline_full\",\"libjvm.so.GraphBuilder::profile_call\",\"libjvm.so.C2Compiler::compile_method\",\"libjvm.so.Compile::Compile\",\"libjvm.so.ParseGenerator::generate\",\"libjvm.so.Parse::Parse\",\"libjvm.so.Parse::do_all_blocks\",\"libjvm.so.Parse::do_one_block\",\"libjvm.so.Parse::do_one_bytecode\",\"libjvm.so.Parse::do_call\",\"libjvm.so.PredictedCallGenerator::generate\",\"libjvm.so.LibraryIntrinsic::generate\",\"libjvm.so.LibraryCallKit::inline_arraycopy\",\"libjvm.so.PhaseGVN::transform_no_reclaim\",\"libjvm.so.AddNode::Ideal\",\"libjvm.so.LoadINode::Opcode\",\"libjvm.so.Compile::Optimize\",\"libjvm.so.Compile::optimize_loops\",\"libjvm.so.PhaseIdealLoop::build_and_optimize\",\"libjvm.so.PhaseIterGVN::optimize\",\"libjvm.so.PhaseIterGVN::transform_old\",\"libjvm.so.PhaseIterGVN::subsume_node\",\"libjvm.so.PhaseIterGVN::remove_globally_dead_node\",\"libjvm.so.Compile::Code_Gen\",\"libjvm.so.PhaseChaitin::Register_Allocate\",\"libjvm.so.PhaseLive::compute\",\"libjvm.so.PhaseLive::add_liveout\",\"libjvm.so.PhaseChaitin::Select\",\"libjvm.so.PhaseIFG::re_insert\",\"libjvm.so.IndexSetIterator::advance_and_next\",\"libjvm.so.Matcher::match\",\"libjvm.so.Matcher::xform\",\"libjvm.so.Matcher::match_tree\",\"libjvm.so.Matcher::Label_Root\",\"libjvm.so.TypeNode::bottom_type\",\"libpthread-2.31.so.pthread_cond_wait@@GLIBC_2.3.2\",\"libpthread-2.31.so.do_futex_wait.constprop.0\",\"libc-2.31.so.__libc_start_main\",\"java.main\",\"libjli.so.JLI_Launch\",\".unknown\",\"libpthread-2.31.so.__pthread_clockjoin_ex\",\"jdk/internal/misc/InnocuousThread.run\",\"java/lang/Thread.run\",\"jdk/internal/ref/CleanerImpl.run\",\"java/lang/ref/ReferenceQueue.remove\",\"java/lang/Object.wait\",\"java/lang/ref/Reference$ReferenceHandler.run\",\"java/lang/ref/Reference.processPendingReferences\",\"java/lang/ref/Reference.waitForReferencePendingList\",\"java/lang/ref/Finalizer$FinalizerThread.run\",\"org/apache/tomcat/util/threads/TaskThread$WrappingRunnable.run\",\"java/util/concurrent/ThreadPoolExecutor$Worker.run\",\"java/util/concurrent/ThreadPoolExecutor.runWorker\",\"java/util/concurrent/ThreadPoolExecutor.getTask\",\"org/apache/tomcat/util/threads/TaskQueue.take\",\"java/util/concurrent/LinkedBlockingQueue.take\",\"java/util/concurrent/locks/LockSupport.park\",\"java/util/concurrent/ScheduledThreadPoolExecutor$DelayedWorkQueue.take\",\"java/util/concurrent/locks/AbstractQueuedSynchronizer$ConditionObject.awaitNanos\",\"libjvm.so.Unsafe_Park\",\"libjvm.so.os::is_interrupted\",\"libjvm.so.Parker::park\",\"libjvm.so.Monitor::lock_without_safepoint_check\",\"libjvm.so.Monitor::ILock\",\"java/util/concurrent/ScheduledThreadPoolExecutor$ScheduledFutureTask.run\",\"java/util/concurrent/FutureTask.runAndReset\",\"java/util/concurrent/Executors$RunnableAdapter.call\",\"org/apache/catalina/core/ContainerBase$ContainerBackgroundProcessor.run\",\"org/apache/catalina/core/ContainerBase$ContainerBackgroundProcessor.processChildren\",\"org/apache/catalina/core/StandardContext.backgroundProcess\",\"org/apache/catalina/core/ContainerBase.backgroundProcess\",\"org/apache/catalina/util/LifecycleBase.fireLifecycleEvent\",\".itable stub\",\"org/apache/tomcat/util/net/NioEndpoint$Poller.run\",\"org/apache/tomcat/util/net/Acceptor.run\",\"org/apache/tomcat/util/net/NioEndpoint.serverSocketAccept\",\"sun/nio/ch/ServerSocketChannelImpl.accept\",\"sun/nio/ch/ServerSocketChannelImpl.accept0\",\"okhttp3/internal/connection/RealCall$AsyncCall.run\",\"okhttp3/internal/connection/RealCall.getResponseWithInterceptorChain$okhttp\",\"okhttp3/internal/http/RealInterceptorChain.proceed\",\"io/opentelemetry/exporter/sender/okhttp/internal/RetryInterceptor.intercept\",\"okhttp3/internal/http/RetryAndFollowUpInterceptor.intercept\",\"okhttp3/internal/http/BridgeInterceptor.intercept\",\"okhttp3/internal/cache/CacheInterceptor.intercept\",\"okhttp3/internal/connection/ConnectInterceptor.intercept\",\"okhttp3/internal/http/CallServerInterceptor.intercept\",\"okhttp3/internal/connection/Exchange.writeRequestHeaders\",\"okhttp3/internal/http2/Http2ExchangeCodec.writeRequestHeaders\",\"okhttp3/internal/http2/Http2ExchangeCodec$Companion.http2HeadersList\",\"java/util/Collections$UnmodifiableCollection.contains\",\"io/opentelemetry/exporter/sender/okhttp/internal/OkHttpGrpcSender$1.onResponse\",\"okhttp3/ResponseBody.bytes\",\"okio/RealBufferedSource.readByteArray\",\"okio/Buffer.writeAll\",\"okhttp3/internal/connection/Exchange$ResponseBodySource.read\",\"okhttp3/internal/connection/Exchange$ResponseBodySource.complete\",\"okhttp3/internal/connection/Exchange.bodyComplete\",\"okhttp3/internal/connection/RealCall.messageDone$okhttp\",\"okhttp3/internal/connection/RealCall.callDone\",\"okhttp3/internal/connection/RealCall.releaseConnectionNoEvents$okhttp\",\"java/util/ArrayList.remove\",\"okhttp3/internal/concurrent/TaskRunner$runnable$1.run\",\"okhttp3/internal/concurrent/TaskRunner.awaitTaskToRun\",\"okhttp3/internal/concurrent/TaskRunner$RealBackend.coordinatorWait\",\"okhttp3/internal/concurrent/TaskRunner.access$runTask\",\"okhttp3/internal/concurrent/TaskRunner.runTask\",\"okhttp3/internal/concurrent/TaskQueue$execute$1.runOnce\",\"okhttp3/internal/http2/Http2Connection$ReaderRunnable.invoke\",\"okhttp3/internal/http2/Http2Reader.nextFrame\",\"okio/RealBufferedSource.require\",\"okio/RealBufferedSource.request\",\"okio/AsyncTimeout$source$1.read\",\"okio/InputStreamSource.read\",\"java/net/SocketInputStream.read\",\"java/net/SocketInputStream.socketRead\",\"java/net/SocketInputStream.socketRead0\",\"libpthread-2.31.so.__libc_recv\",\"okhttp3/internal/http2/Http2Reader.readPing\",\"okhttp3/internal/http2/Http2Connection$ReaderRunnable.ping\",\"okhttp3/internal/concurrent/TaskQueue.schedule\",\"okhttp3/internal/concurrent/TaskRunner.kickCoordinator$okhttp\",\"okhttp3/internal/concurrent/TaskRunner$RealBackend.coordinatorNotify\",\"java/lang/Object.notify\",\"libjvm.so.JVM_MonitorNotify\",\"libjvm.so.ObjectMonitor::notify\",\"java/lang/Thread.setName\",\"java/lang/Thread.setNativeName\",\"libc-2.31.so.prctl\",\"java/util/concurrent/SynchronousQueue.poll\",\"java/util/concurrent/SynchronousQueue$TransferStack.transfer\",\"java/util/concurrent/SynchronousQueue$TransferStack.awaitFulfill\",\"io/opentelemetry/sdk/trace/export/BatchSpanProcessor$Worker.run\",\"java/util/concurrent/ArrayBlockingQueue.poll\",\"java/lang/System.nanoTime\",\"libc-2.31.so.clock_gettime\",\"io/opentelemetry/javaagent/shaded/instrumentation/api/internal/cache/weaklockfree/WeakConcurrentMapCleaner$$Lambda_.run\",\"io/opentelemetry/javaagent/shaded/instrumentation/api/internal/cache/weaklockfree/AbstractWeakConcurrentMap.runCleanup\",\"io/opentelemetry/sdk/metrics/export/PeriodicMetricReader$Scheduled.run\",\"io/opentelemetry/sdk/metrics/export/PeriodicMetricReader$Scheduled.doRun\",\"io/opentelemetry/sdk/metrics/SdkMeterProvider$SdkCollectionRegistration.collectAllMetrics\",\"io/opentelemetry/sdk/metrics/SdkMeterProvider$LeasedMetricProducer.produce\",\"io/opentelemetry/sdk/metrics/SdkMeter.collectAll\",\"io/opentelemetry/sdk/metrics/internal/state/MeterSharedState.collectAll\",\"io/opentelemetry/sdk/metrics/internal/state/CallbackRegistration.invokeCallback\",\"io/opentelemetry/sdk/metrics/InstrumentBuilder$$Lambda_.run\",\"io/opentelemetry/sdk/metrics/InstrumentBuilder.lambda$buildDoubleAsynchronousInstrument$0\",\"io/opentelemetry/javaagent/shaded/instrumentation/runtimemetrics/java8/Cpu$$Lambda_.accept\",\"io/opentelemetry/javaagent/shaded/instrumentation/runtimemetrics/java8/Cpu.lambda$registerObservers$2\",\"io/opentelemetry/javaagent/shaded/instrumentation/runtimemetrics/java8/internal/CpuMethods$$Lambda_.get\",\"io/opentelemetry/javaagent/shaded/instrumentation/runtimemetrics/java8/internal/CpuMethods.lambda$methodInvoker$0\",\"java/lang/reflect/Method.invoke\",\"jdk/internal/reflect/DelegatingMethodAccessorImpl.invoke\",\"jdk/internal/reflect/GeneratedMethodAccessor_.invoke\",\"com/sun/management/internal/OperatingSystemImpl.getProcessCpuLoad\",\"com/sun/management/internal/OperatingSystemImpl$ContainerCpuTicks.getContainerCpuLoad\",\"jdk/internal/platform/cgroupv1/Metrics.getCpuQuota\",\"jdk/internal/platform/cgroupv1/SubSystem.getLongValue\",\"jdk/internal/platform/cgroupv1/SubSystem.getStringValue\",\"jdk/internal/platform/cgroupv1/SubSystem.readStringValue\",\"java/security/AccessController.doPrivileged\",\"jdk/internal/platform/cgroupv1/SubSystem$$Lambda_.run\",\"jdk/internal/platform/cgroupv1/SubSystem.lambda$readStringValue$0\",\"java/nio/file/Files.newBufferedReader\",\"java/nio/file/Files.newInputStream\",\"java/nio/file/spi/FileSystemProvider.newInputStream\",\"java/nio/file/Files.newByteChannel\",\"sun/nio/fs/UnixFileSystemProvider.newByteChannel\",\"sun/nio/fs/UnixChannelFactory.newFileChannel\",\"sun/nio/fs/UnixChannelFactory.open\",\"io/opentelemetry/exporter/otlp/metrics/OtlpGrpcMetricExporter.export\",\"io/opentelemetry/exporter/internal/otlp/metrics/MetricsRequestMarshaler.create\",\"io/opentelemetry/exporter/internal/otlp/metrics/ResourceMetricsMarshaler.create\",\"io/opentelemetry/exporter/internal/otlp/ResourceMarshaler.create\"],\"levels\":[[0,5084230000000,0,0],[0,3319940000000,0,94,0,103790000000,0,101,0,103800000000,0,98,0,103800000000,0,93,0,103780000000,0,88,0,622480000000,622480000000,4,0,103830000000,103830000000,87,0,207610000000,207610000000,86,0,103880000000,0,16,0,103760000000,0,11,0,103760000000,10000000,5,0,103800000000,0,1],[0,103790000000,0,91,0,103710000000,0,188,0,103760000000,0,184,0,518660000000,0,103,0,103750000000,0,126,0,103730000000,0,125,0,2282540000000,0,102,0,103790000000,0,96,0,103800000000,0,99,0,103800000000,0,94,0,103780000000,0,89,933920000000,103880000000,0,17,0,103760000000,0,12,10000000,103750000000,0,6,0,103800000000,0,2],[0,103790000000,0,104,0,103710000000,0,189,0,10000000,0,186,0,103750000000,0,185,0,518660000000,0,104,0,103750000000,0,127,0,103730000000,0,6,0,2282540000000,0,103,0,103790000000,0,96,0,103800000000,0,100,0,103800000000,0,95,0,103780000000,0,90,933920000000,103880000000,0,18,0,103760000000,0,13,10000000,103750000000,0,7,0,103800000000,0,3],[0,30000000,0,116,0,103760000000,0,105,0,103710000000,0,96,0,10000000,10000000,187,0,103750000000,0,110,0,311120000000,0,105,0,207520000000,0,154,0,20000000,0,130,0,103750000000,0,127,0,103730000000,0,7,0,2282540000000,10000000,104,0,103790000000,0,97,0,103800000000,103800000000,86,0,103800000000,0,96,0,103780000000,0,91,933920000000,103830000000,0,25,0,50000000,0,19,0,103760000000,0,14,10000000,103750000000,0,8,0,103800000000,103800000000,4],[0,30000000,0,117,0,103760000000,0,109,0,103710000000,0,96,10000000,103750000000,0,14,0,103720000000,0,109,0,207400000000,0,181,0,103790000000,0,157,0,103730000000,0,155,0,10000000,0,143,0,10000000,0,131,0,103750000000,0,128,0,103730000000,0,8,10000000,10000000,0,116,0,2282520000000,0,105,0,103790000000,103790000000,86,103800000000,103800000000,0,97,0,103780000000,103780000000,92,933920000000,90000000,0,29,0,103740000000,0,26,0,50000000,0,20,0,103760000000,0,15,10000000,103750000000,0,9],[0,30000000,0,118,0,103760000000,0,109,0,103710000000,0,97,10000000,103750000000,0,15,0,103720000000,0,109,0,207400000000,0,182,0,103790000000,0,158,0,103730000000,0,156,0,10000000,0,144,0,10000000,0,132,0,103750000000,0,128,0,103730000000,0,9,10000000,10000000,0,117,0,207490000000,0,109,0,2075030000000,0,106,207590000000,103800000000,103800000000,4,1037700000000,90000000,0,30,0,103740000000,0,27,0,50000000,0,21,0,103760000000,103760000000,4,10000000,103750000000,103750000000,10],[0,30000000,0,190,0,103760000000,0,110,0,103710000000,103710000000,86,10000000,103750000000,103750000000,4,0,103720000000,0,110,0,207400000000,0,183,0,10000000,0,178,0,103780000000,0,159,0,103730000000,0,97,0,10000000,0,145,0,10000000,0,133,0,103750000000,0,129,0,103730000000,103730000000,10,10000000,10000000,0,118,0,207490000000,0,109,0,2075030000000,0,106,1349090000000,50000000,0,53,0,30000000,0,33,0,10000000,0,31,0,103740000000,103740000000,28,0,50000000,0,22],[0,30000000,0,191,0,103760000000,0,14,207470000000,103720000000,0,14,0,207400000000,0,14,0,10000000,0,179,0,103780000000,0,160,0,103730000000,0,97,0,10000000,0,146,0,10000000,0,132,0,103750000000,103750000000,28,103740000000,10000000,0,119,0,103730000000,0,13,0,103760000000,0,110,0,2075030000000,0,107,1349090000000,50000000,0,54,0,30000000,0,34,0,10000000,10000000,32,103740000000,10000000,10000000,4,0,30000000,30000000,24,0,10000000,10000000,23],[0,10000000,0,222,0,20000000,0,192,0,103760000000,0,15,207470000000,103720000000,0,15,0,207400000000,0,15,0,10000000,0,91,0,103780000000,0,160,0,103730000000,103730000000,4,0,10000000,0,147,0,10000000,0,134,207490000000,10000000,0,120,0,103730000000,0,108,0,103760000000,0,14,0,2075030000000,0,13,1349090000000,30000000,0,74,0,10000000,0,67,0,10000000,0,55,0,30000000,0,35],[0,10000000,0,223,0,20000000,0,193,0,103760000000,103760000000,4,207470000000,10000000,0,111,0,103710000000,103710000000,4,0,207400000000,207400000000,4,0,10000000,10000000,180,0,103780000000,0,161,103730000000,10000000,0,148,0,10000000,0,132,207490000000,10000000,0,120,0,103730000000,0,15,0,103760000000,0,15,0,2075030000000,0,108,1349090000000,10000000,0,81,0,20000000,0,75,0,10000000,0,68,0,10000000,0,56,0,30000000,0,36],[0,10000000,0,224,0,20000000,0,194,311230000000,10000000,0,113,311120000000,10000000,0,170,0,103770000000,0,162,103730000000,10000000,0,149,0,10000000,0,135,207490000000,10000000,0,120,0,10000000,0,111,0,103720000000,103720000000,86,0,20000000,0,111,0,103740000000,103740000000,4,0,2075030000000,0,15,1349090000000,10000000,0,82,0,10000000,0,78,0,10000000,0,76,0,10000000,0,69,0,10000000,0,57,0,20000000,0,43,0,10000000,0,37],[0,10000000,10000000,225,0,20000000,10000000,195,311230000000,10000000,0,114,311120000000,10000000,0,171,0,103770000000,0,163,103730000000,10000000,0,150,0,10000000,0,132,207490000000,10000000,0,121,0,10000000,10000000,113,103720000000,10000000,0,113,0,10000000,10000000,112,103740000000,2075030000000,2075030000000,86,1349090000000,10000000,0,83,0,10000000,0,79,0,10000000,10000000,77,0,10000000,0,70,0,10000000,0,58,0,20000000,0,44,0,10000000,0,38],[20000000,10000000,0,196,311230000000,10000000,10000000,115,311120000000,10000000,0,172,0,103770000000,0,164,103730000000,10000000,0,151,0,10000000,0,136,207490000000,10000000,0,122,103730000000,10000000,0,114,3527870000000,10000000,0,84,0,10000000,10000000,80,10000000,10000000,0,71,0,10000000,0,59,0,20000000,0,45,0,10000000,0,39],[20000000,10000000,0,197,622360000000,10000000,0,173,0,103770000000,0,165,103730000000,10000000,0,152,0,10000000,0,132,207490000000,10000000,0,123,103730000000,10000000,10000000,115,3527870000000,10000000,10000000,85,20000000,10000000,0,72,0,10000000,0,60,0,10000000,0,47,0,10000000,10000000,46,0,10000000,0,40],[20000000,10000000,0,198,622360000000,10000000,0,174,0,103770000000,0,166,103730000000,10000000,10000000,153,0,10000000,0,137,207490000000,10000000,10000000,124,3631640000000,10000000,10000000,73,0,10000000,0,55,0,10000000,0,48,10000000,10000000,0,41],[20000000,10000000,0,199,622360000000,10000000,0,175,0,103770000000,0,166,103740000000,10000000,0,132,3839150000000,10000000,0,56,0,10000000,0,49,10000000,10000000,10000000,42],[20000000,10000000,0,200,622360000000,10000000,0,176,0,103770000000,0,167,103740000000,10000000,0,138,3839150000000,10000000,0,57,0,10000000,0,50],[20000000,10000000,0,201,622360000000,10000000,10000000,177,0,103770000000,0,168,103740000000,10000000,0,139,3839150000000,10000000,0,58,0,10000000,0,51],[20000000,10000000,0,202,622370000000,103760000000,103760000000,169,0,10000000,10000000,86,103740000000,10000000,0,140,3839150000000,10000000,0,59,0,10000000,0,47],[20000000,10000000,0,203,829880000000,10000000,0,141,3839150000000,10000000,0,60,0,10000000,0,48],[20000000,10000000,0,204,829880000000,10000000,10000000,142,3839150000000,10000000,0,61,0,10000000,0,49],[20000000,10000000,0,205,4669040000000,10000000,0,55,0,10000000,0,50],[20000000,10000000,0,206,4669040000000,10000000,0,56,0,10000000,0,51],[20000000,10000000,0,207,4669040000000,10000000,0,57,0,10000000,10000000,52],[20000000,10000000,0,208,4669040000000,10000000,0,58],[20000000,10000000,0,209,4669040000000,10000000,0,59],[20000000,10000000,0,210,4669040000000,10000000,0,60],[20000000,10000000,0,211,4669040000000,10000000,0,55],[20000000,10000000,0,212,4669040000000,10000000,0,56],[20000000,10000000,0,213,4669040000000,10000000,0,57],[20000000,10000000,0,214,4669040000000,10000000,0,58],[20000000,10000000,0,215,4669040000000,10000000,0,59],[20000000,10000000,0,215,4669040000000,10000000,0,60],[20000000,10000000,0,216,4669040000000,10000000,0,55],[20000000,10000000,0,217,4669040000000,10000000,0,56],[20000000,10000000,0,218,4669040000000,10000000,0,57],[20000000,10000000,0,218,4669040000000,10000000,0,58],[20000000,10000000,0,219,4669040000000,10000000,0,59],[20000000,10000000,0,220,4669040000000,10000000,0,60],[20000000,10000000,0,220,4669040000000,10000000,0,55],[20000000,10000000,10000000,221,4669040000000,10000000,0,56],[4669070000000,10000000,0,57],[4669070000000,10000000,0,58],[4669070000000,10000000,0,59],[4669070000000,10000000,0,60],[4669070000000,10000000,0,55],[4669070000000,10000000,0,56],[4669070000000,10000000,0,57],[4669070000000,10000000,0,58],[4669070000000,10000000,0,59],[4669070000000,10000000,0,60],[4669070000000,10000000,0,62],[4669070000000,10000000,0,63],[4669070000000,10000000,0,64],[4669070000000,10000000,0,65],[4669070000000,10000000,0,64],[4669070000000,10000000,10000000,66]],\"numTicks\":5084230000000,\"maxSelf\":2075030000000}"
	flamebearer := FlameBearer{}
	err := json.Unmarshal([]byte(data), &flamebearer)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(wrapToFlame(flamebearer))
}
