const Capability = ({ capabilitySpec }: { capabilitySpec: string[] }) => {
  return (
    <div className="flex flex-wrap gap-2">
      {capabilitySpec.map((type, index) => (
        <span
          key={index}
          className="text-xs bg-blue-50 text-blue-600 px-1 py-1 rounded font-medium border border-blue-200 hover:bg-blue-100 transition-colors duration-200 dark:bg-blue-900/20 dark:text-blue-400 dark:border-blue-800 dark:hover:bg-blue-900/30"
        >
          {type}
        </span>
      ))}
    </div>
  )
}
export default Capability
