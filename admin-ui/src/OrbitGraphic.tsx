function OrbitGraphic() {
  return (
    <div className="orbit-field" aria-hidden="true">
      <svg viewBox="0 0 1440 900" preserveAspectRatio="xMidYMid slice">
        <path
          className="orbit-line orbit-line-primary"
          d="M-140 548C180 182 790 76 1336 288c246 93 275 284 22 458"
        />
        <path
          className="orbit-line orbit-line-secondary"
          d="M-180 786c382-290 900-390 1368-280 212 50 326 166 274 274"
        />
        <circle className="orbit-halo" cx="1218" cy="264" r="24" />
        <circle className="orbit-point" cx="1218" cy="264" r="8" />
      </svg>
    </div>
  )
}

export default OrbitGraphic
