const fs = require('fs')
const path = require('path')
const http = require('http');

const folder = process.argv[2]
const projectId = process.argv[3]

if (!folder) {
  console.log('Need folder')
  process.exit(1)
}

if (!projectId) {
  console.log('Need projectId')
  process.exit(1)
}

let reportFiles = [
  'report1.json', 'report2.json', 'report3.json', 'report4.json', 'report5.json',
  'report6.json', 'report7.json', 'report8.json', 'report9.json', 'report1.json', 
  'report2.json', 'report3.json', 'report4.json', 'report5.json', 'report6.json', 
  'report7.json', 'report8.json', 'report9.json'
]

reportFiles = arrayShuffle(reportFiles)

reportFiles.push(`report3.json`)

createData()

async function createData () {
  let n = 0
  let h = 0
  const MONTH = (new Date()).getMonth()
  reportFiles.forEach(async fileName => {

    const rf = path.join(__dirname, folder, fileName)

    try {
      if (!rf) {
        return
      }

      n++
      if (n > 27) {
        console.log('maximum reached skipping...')
        return
      }

      h++
      if (h >= 23) {
        h = 1
      }

      console.log(rf)
      const content = fs.readFileSync(rf, 'utf8')
      const data = JSON.parse(content)
      const date = new Date()
      date.setMonth(MONTH)
      date.setDate(n)
      date.setHours(h)
      data.date = date.toISOString()
      const status = await doPost(data)
      console.log('done: ' + status)
    } catch (e) {
      console.log(e)
    }
  })
}

function doPost (data) {
  return new Promise((resolve, reject) => {
    const postData = JSON.stringify(data)

    const options = {
      hostname: 'localhost',
      port: 3000,
      path: `/api/projects/${projectId}/ingest`,
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-cache',
        'Content-Length': postData.length
      }
    }

    req = http.request(options, res => {
      res.setEncoding('utf8')
      res.on('end', () => {
        if (!res.complete) {
          console.error(
            'The connection was terminated while the message was still being sent');
          reject(new Error('Incomplete'))
        } else {
          resolve(res.statusCode)
        }
      });
    })

    req.on('error', function (e) {
      reject(e)
    })

    req.write(postData);
    req.end()
  })
}

function arrayShuffle(array) {
  return shuffleSelf(copyArray(array));
}

function copyArray(source, array) {
  var index = -1,
      length = source.length;

  array || (array = Array(length));
  while (++index < length) {
    array[index] = source[index];
  }
  return array;
}

function shuffleSelf(array, size) {
  var index = -1,
      length = array.length,
      lastIndex = length - 1;

  size = size === undefined ? length : size;
  while (++index < size) {
    var rand = baseRandom(index, lastIndex),
        value = array[rand];

    array[rand] = array[index];
    array[index] = value;
  }
  array.length = size;
  return array;
}

function baseRandom(lower, upper) {
  return lower + Math.floor(Math.random() * (upper - lower + 1));
}

function getRandomInt (max) {
  return Math.floor(Math.random() * Math.floor(max))
}
