function FullTrace() {
  return (
    <div className="text-xs" style={{ height: 'calc(100vh - 120px)' }}>
      <iframe src={'/jaeger/search'} width="100%" height="100%" frameBorder={0}></iframe>
    </div>
  )
}

export default FullTrace
