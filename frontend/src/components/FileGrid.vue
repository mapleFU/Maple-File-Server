<template>
  <div class="file-grid">
    <h2> {{currentDir}} </h2>
    <v-container v-bind="{ [`grid-list-${size}`]: true }" fluid>
      <v-layout row wrap>
        <v-flex
          v-for="file in file_list"
          :key="file"
          xs4
        >
          <template v-if="file.fileType === 'file'">
            <File v-bind:fileName='file.fileName'> </File>
          </template>
          <template v-else>
            <Dictory v-bind:fileName='file.fileName' ></Dictory>
          </template>

        </v-flex>
      </v-layout>
    </v-container>
  </div>
</template>

<script>
import File from './File'
import Dictory from './Dictory'

export default {
  name: "file-grid",
  data: () => ({
    file_list: [{
      fileName: "mwish",
      fileType: "file",
    },],
    currentDir: "root",
    size: 'sm',
  }),
  components: {
    File, Dictory
  }
}
</script>

<style scoped>

</style>
