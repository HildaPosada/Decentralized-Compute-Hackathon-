import React from 'react'

function NodesList({ nodes }) {
  return (
    <div className="card">
      <h2>🖥️ Worker Nodes ({nodes.length})</h2>
      <div className="node-list">
        {nodes.length === 0 ? (
          <p style={{ color: '#999', textAlign: 'center', padding: '20px' }}>
            No nodes registered yet
          </p>
        ) : (
          nodes.map((node) => (
            <div key={node.id} className="node-item">
              <div className="node-header">
                <div className="node-name">{node.name}</div>
                <span className={`status-badge status-${node.status}`}>
                  {node.status}
                </span>
              </div>
              <div className="node-details">
                <div className="detail-item">
                  <span className="detail-label">Region:</span>
                  <span>{node.region || 'unknown'}</span>
                </div>
                <div className="detail-item">
                  <span className="detail-label">Reputation:</span>
                  <span>{node.reputation_score.toFixed(1)}/100</span>
                </div>
                <div className="detail-item">
                  <span className="detail-label">CPU:</span>
                  <span>{node.cpu_cores} cores</span>
                </div>
                <div className="detail-item">
                  <span className="detail-label">Memory:</span>
                  <span>{node.memory_gb} GB</span>
                </div>
                <div className="detail-item">
                  <span className="detail-label">Jobs Run:</span>
                  <span>{node.total_jobs_run}</span>
                </div>
                <div className="detail-item">
                  <span className="detail-label">Success Rate:</span>
                  <span>
                    {node.total_jobs_run > 0
                      ? ((node.successful_jobs_run / node.total_jobs_run) * 100).toFixed(1)
                      : 0}%
                  </span>
                </div>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  )
}

export default NodesList
