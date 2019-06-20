const FFI = require('ffi')
const Ref = require('ref')
var Array = require('ref-array')
var Struct = require('ref-struct')
var sprintf = require('sprintf-js')
const hex2ascii = require('hex2ascii')

// define object GoString to map:
// C type struct { const char *p; GoInt n; }
var GoString = Struct({
  p: 'string',
  n: 'longlong'
})

let ucharArray = Array(Ref.types.uchar)
const hwSharedLibPath = './go-node-ffi'

const hhh = FFI.Library(hwSharedLibPath, {
  'GetSeed': ['int', [ucharArray]],
  'GetData': ['int', [ucharArray, 'pointer', GoString]]
})

function toHex (charArray, len) {
  var converted = ''
  var str = ''
  for (var i = 0; i < len; i++) {
    str = sprintf.sprintf('%02x', charArray[i])
    converted = converted + str
  }
  return converted
}

const getSeed = async () => {
  var seedArray = ucharArray(65)
  await hhh.GetSeed(seedArray)
  var seed = hex2ascii(toHex(seedArray, 64))

  // todo: for TEST only
  // console.log(seed)

  return seed
}

const getData = async (tt) => {
  var dataArray = ucharArray(256)
  var lenP = Ref.alloc('int')

  let str = new GoString()
  str['p'] = tt
  str['n'] = tt.length

  await hhh.GetData(dataArray, lenP, str)

  var data = hex2ascii(toHex(dataArray, lenP.deref()))

  return data
}

module.exports = {
  getSeed,
  getData
}
