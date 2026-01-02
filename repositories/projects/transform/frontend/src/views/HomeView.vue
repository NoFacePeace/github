<script setup lang="ts">
import Header from '../components/Header.vue'
import Aside from '../components/Aside.vue';
import Main from '../components/Main.vue'
</script>
<template>
  <Header></Header>
  <el-container class="h-screen">
    <el-aside>
      <Aside></Aside>
    </el-aside>
    <el-main>
      <Main></Main>
    </el-main>
  </el-container>
</template>
<style scoped></style>