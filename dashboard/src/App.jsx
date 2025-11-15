import React, { useState, useEffect } from 'react'
import axios from 'axios'
import StatsCards from './components/StatsCards'
import NodesList from './components/NodesList'
import JobsList from './components/JobsList'

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

function App() {
  const [stats, setStats] = useState(null)
  const [nodes, setNodes] = useState([])
  const [jobs, setJobs] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [lastUpdate, setLastUpdate] = useState(new Date())

  const fetchData = async () => {
    try {
      const [statsRes, nodesRes, jobsRes] = await Promise.all([
        axios.get(`${API_URL}/stats`),
        axios.get(`${API_URL}/api/v1/nodes`),
        axios.get(`${API_URL}/api/v1/jobs`),
      ])

      setStats(statsRes.data)
      setNodes(nodesRes.data.nodes || [])
      setJobs(jobsRes.data.jobs || [])
      setError(null)
      setLastUpdate(new Date())
    } catch (err) {
      setError(err.message)
      console.error('Error fetching data:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
    const interval = setInterval(fetchData, 5000) // Refresh every 5 seconds
    return () => clearInterval(interval)
  }, [])

  if (loading) {
    return <div className="loading">Loading DistributeAI Dashboard...</div>
  }

  return (
    <div className="container">
      <div className="header">
        <h1>⚡ DistributeAI</h1>
        <p>
          Decentralized Compute Network - Powered by the People
          <span className="refresh-badge">
            Last updated: {lastUpdate.toLocaleTimeString()}
          </span>
        </p>
      </div>

      {error && (
        <div className="error">
          <strong>Error:</strong> {error}
        </div>
      )}

      {stats && <StatsCards stats={stats} />}

      <div className="grid">
        <NodesList nodes={nodes} />
        <JobsList jobs={jobs} />
      </div>
    </div>
  )
}

export default App
