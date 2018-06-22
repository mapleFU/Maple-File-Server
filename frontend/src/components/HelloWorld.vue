<template>
  <div class="hello">

    <h1>{{ msg }}</h1>
    <div v-if='fileNums > 0'>
      <File :fileName="'nmsl'"></File>
      <Dictory></Dictory>
    </div>

  </div>
</template>

<script>
import File from './File'
import Dictory from './Dictory'

export default {
  name: 'HelloWorld',
  data () {
    return {
      fileNums: 1,
      current_directory: '',
      msg: '你说你🐎呢'
    }
  },
  components: {
    File, Dictory
  },
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
h1, h2 {
  font-weight: normal;
}
ul {
  list-style-type: none;
  padding: 0;
}
li {
  display: inline-block;
  margin: 0 10px;
}
a {
  color: #42b983;
}
</style>
