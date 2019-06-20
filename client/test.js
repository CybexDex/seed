const { getSeed, getData } = require('./go-node-ffi')

const testGetSeed = async () => {
  let seed = await getSeed()
  console.log()  
  console.log('SEED', seed)
}

const testGetData = async () => {
  let data = await getData('BTC')
  console.log()
  console.log('BTC', data)

  data = await getData('ETH')
  console.log()
  console.log('ETH', data)

  data = await getData('CYB')
  console.log()
  console.log('CYB', data)
}

testGetSeed()
testGetData()
