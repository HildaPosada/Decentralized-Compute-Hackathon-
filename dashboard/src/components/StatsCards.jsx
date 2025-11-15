import React from 'react'

function StatsCards({ stats }) {
  return (
    <div className="stats-grid">
      <div className="stat-card">
        <h3>Total Nodes</h3>
        <div className="value">{stats.nodes.total}</div>
        <div style={{ fontSize: '0.9rem', color: '#666', marginTop: '10px' }}>
          {stats.nodes.online} online · {stats.nodes.busy} busy
        </div>
      </div>

      <div className="stat-card">
        <h3>Total CPU Cores</h3>
        <div className="value">{stats.resources.total_cpu_cores}</div>
        <div style={{ fontSize: '0.9rem', color: '#666', marginTop: '10px' }}>
          {stats.resources.total_memory_gb} GB RAM
        </div>
      </div>

      <div className="stat-card">
        <h3>Jobs Completed</h3>
        <div className="value">{stats.jobs.completed}</div>
        <div style={{ fontSize: '0.9rem', color: '#666', marginTop: '10px' }}>
          {stats.jobs.running} running · {stats.jobs.failed} failed
        </div>
      </div>

      <div className="stat-card">
        <h3>Total Jobs</h3>
        <div className="value">{stats.jobs.total}</div>
        <div style={{ fontSize: '0.9rem', color: '#666', marginTop: '10px' }}>
          All-time submissions
        </div>
      </div>
    </div>
  )
}

export default StatsCards
