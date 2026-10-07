import React, { useEffect, useState } from 'react'
import { Pane, Table, Heading, Text, Button, Spinner, Alert, Tab, Icon, Textarea } from 'evergreen-ui'

export default function TestDataPage () {
  const [files, setFiles] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [selected, setSelected] = useState(null)
  const [content, setContent] = useState('')

  useEffect(() => {
    setLoading(true)
    fetch('/api/testdata/list/')
      .then(r => r.json())
      .then(data => { setFiles(data.files || []); })
      .catch(e => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  const loadFile = (p) => {
    setSelected(p)
    setContent('')
    fetch(`/api/testdata/file/?p=${encodeURIComponent(p)}`)
      .then(async r => {
        if (!r.ok) throw new Error(await r.text())
        return r.text()
      })
      .then(txt => setContent(txt))
      .catch(e => setError(e.message))
  }

  const saveToRun = () => {
    if (!selected) return
    if (selected.endsWith('.proto')) {
      localStorage.setItem('run.proto_file', content)
    } else if (selected.endsWith('.json')) {
      localStorage.setItem('run.data', content)
    }
  }

  return (
    <Pane padding={16}>
      <Heading size={700} marginBottom={12}>Test Data</Heading>
      {error && <Alert intent='danger' title='Error' marginBottom={12}>{error}</Alert>}
      {loading && <Spinner />}
      <Pane display='flex'>
        <Pane flex={1} marginRight={16}>
          <Table>
            <Table.Head>
              <Table.TextHeaderCell>File</Table.TextHeaderCell>
              <Table.TextHeaderCell>Type</Table.TextHeaderCell>
              <Table.TextHeaderCell flex={0}>Action</Table.TextHeaderCell>
            </Table.Head>
            <Table.Body height={480}>
              {files.map(f => (
                <Table.Row key={f.path} isSelectable onSelect={() => loadFile(f.path)} intent={selected === f.path ? 'success' : 'none'}>
                  <Table.TextCell>{f.path}</Table.TextCell>
                  <Table.TextCell>{f.type}</Table.TextCell>
                  <Table.Cell flex={0}>
                    <Button size='small' onClick={() => loadFile(f.path)}>²é¿´</Button>
                  </Table.Cell>
                </Table.Row>
              ))}
            </Table.Body>
          </Table>
        </Pane>
        <Pane flex={1}>
          {selected ? (
            <Pane>
              <Heading size={600} marginBottom={8}>{selected}</Heading>
              <Textarea value={content} onChange={e => setContent(e.target.value)} height={360} marginBottom={12} />
              <Button appearance='primary' onClick={saveToRun} marginRight={8}>Save to Run Page</Button>
              <Text size={300} color='muted'>proto will be written to run.proto_file; json will be written to run.data</Text>
            </Pane>
          ) : <Text>Select a file to view content</Text>}
        </Pane>
      </Pane>
    </Pane>
  )
}
