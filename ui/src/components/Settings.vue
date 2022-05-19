<template>
  <div class="setting">
    <h3 class="title">{{ $t('Settings') }}</h3>
    <ul style="padding: 0">
      <li
        v-for="(i, index) in settings"
        :key="index"
        class="item"
      >
        <el-button
          type="text"
          class="item-button"
          :disabled="!i.disabled()"
          :class="'icon ' + i.icon"
          @click.stop="i.click && i.click()"
        >
          {{ i.title }}
        </el-button>
        <div v-if="i.content" class="content">
          <el-button
            v-for="(item, key) of i.content"
            :key="key"
            class="content-item"
            type="text"
            :disabled="i.disabled()"
            @click="i.itemClick && i.itemClick(item.keys)"
          >
            {{ item.name }}
          </el-button>
        </div>
      </li>
    </ul>
  </div>
</template>

<script>
export default {
  name: 'Settings',
  props: {
    settings: {
      type: Array,
      default: () => []
    }
  }
}
</script>

<style scoped>
.setting {
  padding: 24px 24px;
}

.title {
  text-align: left;
  padding-left: 12px;
  font-size: 18px;
  color: #000;
}

.item {
  color: rgba(0, 0, 0, 0.65);
  font-size: 14px;
  list-style-type: none;
  cursor: pointer;
  border-radius: 2px;
  line-height: 14px;
}

.item-button {
  width: 100%;
  text-align: left;
  color: #303133;
}

.item-button:hover {
  color: #fff;
  background: rgba(0, 0, 0, .3);
}

.item-button.is-disabled {
  color: #C0C4CC;
}

.content {
  padding: 4px 6px 4px 20px;
}

.content-item {
  display: block;
  width: 100%;
  margin-left: 0;
  padding: 4px 0;
  overflow: hidden;
  text-align: left;
  white-space: nowrap;
  text-overflow: ellipsis;
  color: black;
}

.content-item.is-disabled {
  color: #C0C4CC;
}

.content-item:hover {
  color: white;
  border-radius: 2px;
  background: rgba(0, 0, 0, .3);
}
</style>
