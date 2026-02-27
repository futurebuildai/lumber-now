interface Props {
  children: React.ReactNode
}

export default function DeviceFrame({ children }: Props) {
  return (
    <div className="relative mx-auto" style={{ width: 280, height: 500 }}>
      {/* Phone frame */}
      <div className="absolute inset-0 bg-foreground rounded-[2.5rem] shadow-xl" />
      {/* Screen area */}
      <div className="absolute inset-[3px] bg-background rounded-[2.3rem] overflow-hidden">
        {/* Notch */}
        <div className="absolute top-0 left-1/2 -translate-x-1/2 w-24 h-6 bg-foreground rounded-b-2xl z-10" />
        {/* Content */}
        <div className="absolute inset-0 pt-8 overflow-hidden">
          {children}
        </div>
      </div>
    </div>
  )
}
