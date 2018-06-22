<template>
  <div class="fs-file">
    <img v-bind:src="imageUri" height="50" width="50"/>
    <span> {{fileName}} </span>
  </div>
</template>

<script>
  import FileImf from '@/assets/text.png'
  export default {
      name: "file",
      data: function () {
        return {
          fileName: "",
          imageUri: FileImf,
        }
      },
      props:['fileName'],
  }
</script>

<style scoped>

</style>
