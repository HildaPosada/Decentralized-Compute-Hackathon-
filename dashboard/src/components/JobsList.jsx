import React from 'react'

function JobsList({ jobs }) {
  // Show only the 10 most recent jobs
  const recentJobs = jobs.slice(0, 10)

  return (
    <div className="card">
      <h2>📦 Recent Jobs ({jobs.length})</h2>
      <div className="job-list">
        {recentJobs.length === 0 ? (
          <p style={{ color: '#999', textAlign: 'center', padding: '20px' }}>
            No jobs submitted yet
          </p>
        ) : (
          recentJobs.map((job) => (
            <div key={job.id} className="job-item">
              <div className="job-header">
                <div className="job-name">{job.name}</div>
                <span className={`status-badge status-${job.status}`}>
                  {job.status}
                </span>
              </div>
              {job.description && (
                <p style={{ fontSize: '0.9rem', color: '#666', marginBottom: '10px' }}>
                  {job.description}
                </p>
              )}
              <div className="job-details">
                <div className="detail-item">
                  <span className="detail-label">Image:</span>
                  <span style={{ fontSize: '0.85rem' }}>{job.docker_image}</span>
                </div>
                <div className="detail-item">
                  <span className="detail-label">Submitted:</span>
                  <span>{new Date(job.submitted_at).toLocaleString()}</span>
                </div>
                <div className="detail-item">
                  <span className="detail-label">CPU/Memory:</span>
                  <span>{job.required_cpu}c / {job.required_memory}GB</span>
                </div>
                <div className="detail-item">
                  <span className="detail-label">Verification:</span>
                  <span>{job.consensus}/{job.redundancy} consensus</span>
                </div>
              </div>
              {job.result && (
                <div style={{
                  marginTop: '10px',
                  padding: '10px',
                  background: '#f5f5f5',
                  borderRadius: '4px',
                  fontSize: '0.85rem',
                  fontFamily: 'monospace',
                  maxHeight: '100px',
                  overflow: 'auto'
                }}>
                  {job.result}
                </div>
              )}
            </div>
          ))
        )}
      </div>
    </div>
  )
}

export default JobsList
