<script setup lang="ts">
import data from '../data/aside.json'
</script>

<template>
    <el-scrollbar>
        <section v-for="(group, gkey) in data" :key="gkey">
            <p class="pl-5 pt-5">
                            <el-text size="large" tag="b">{{ group.text }}</el-text>
            </p>
            <el-menu>
                <el-menu-item v-for="(item, ikey) in group.children" :index="ikey" :key="ikey" :group="gkey">
                    {{ item.text }}
                </el-menu-item>
            </el-menu>
        </section>
    </el-scrollbar>
</template>

<style scoped></style>
