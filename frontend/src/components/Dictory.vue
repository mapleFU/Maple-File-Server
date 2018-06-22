<template>
  <div class="fs-folder">
    <img v-bind:src="imageUri" height="50" width="50"/>
    <span>{{fileName}}</span>
  </div>
</template>

<script>
import DicImf from '@/assets/folder.png'
export default {
  name: "dictory",
  data: function () {
    return {
      fileName: "",
      imageUri: DicImf,
    }
  },
  props:['fileName'],
}
</script>

<style scoped>

</style>
