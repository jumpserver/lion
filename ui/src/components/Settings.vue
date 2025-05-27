<template>
  <div class="setting">
    <h3 class="title">{{ title }}</h3>
    <ul style="padding: 0">
      <li
        v-for="(i, index) in settings"
        :key="index"
        class="item"
      >
        <el-button
          type="text"
          class="item-button"
          :disabled="i.disabled()"
          :class="'icon ' + i.icon"
          @click.stop="i.click && itemClick(i)"
        >
          {{ i.title }}
          <span v-if="i.content && Object.keys(i.content).length > 0 && i.content[0].faIcon">
            ({{ Object.keys(i.content).length }})
          </span>
        </el-button>
        <div v-if="i.content" class="content">
          <div
            v-for="(item, key) of i.content"
            :key="key"
            style="padding-bottom: 1px;"
          >
            <el-tooltip v-if="item.faIcon" class="item" effect="dark" :content="item.iconTip" placement="top-start">
              <FontAwesomeIcon :icon="item.faIcon" />
            </el-tooltip>
            <span v-if="i.itemActions" style="padding-left: 5px;">
              {{ item.name }}
            </span>
            <el-button
              v-else
              :key="key"
              class="content-item"
              type="text"
              :disabled="i.disabled()"
              @click="i.itemClick && i.itemClick(item.keys)"
            >
              {{ item.name }}
            </el-button>
            <span v-if="i.itemActions">
              <span
                v-for="(action, idx) of i.itemActions"
                v-show="!action.hidden(item)"
                :key="idx"
                style="float: right"
                @click.stop="action.click(item)"
              >
                <el-tooltip v-if="action.faIcon" class="item" effect="dark" :content="action.tipText" placement="top-start">
                  <FontAwesomeIcon :icon="action.faIcon" :style="action.style" />
                </el-tooltip>
              </span>
            </span>
          </div>
        </div>
      </li>
      <li class="item">
        <slot />
      </li>
    </ul>
  </div>
</template>

<script>
export default {
  name: 'Settings',
  props: {
    title: {
      type: String,
      required: true
    },
    settings: {
      type: Array,
      default: () => []
    }
  },
  methods: {
    itemClick(item) {
      this.$parent.show = false
      item.click()
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
  color: #ffffff;
}

.item {
  color: rgb(255, 255, 255);
  font-size: 14px;
  list-style-type: none;
  cursor: pointer;
  border-radius: 2px;
  line-height: 14px;
}

.item-button {
  padding-left: 10px;
  width: 100%;
  text-align: left;
  color: #ffffff;
}

.item-button.is-disabled {
  color: rgb(0, 0, 0, 0.5);
}

.item-button.is-disabled:hover {
  color: rgb(0, 0, 0, 0.5);
  background: none;
}

.item-button:hover {
  background: rgba(0, 0, 0, .5);
}

.content {
  padding: 4px 6px 4px 25px;
}

.content-item {
  font-size: 13px;
  white-space: nowrap;
  text-overflow: ellipsis;
  overflow: hidden;
  padding: 4px 0;
  color: #ffffff;
  margin-left: 0;
  display: block;
  width: 100%;
  text-align: left;
}

.content-item:hover {
  border-radius: 2px;
  background: rgba(0, 0, 0, .5);
}

.item {
  color: rgb(250, 247, 247);
  font-size: 14px;
  list-style-type: none;
  cursor: pointer;
  border-radius: 2px;
  line-height: 14px;
}

.item-button {
  padding-left: 10px;
  width: 100%;
  text-align: left;
  color: #faf7f7;
}

.item-button.is-disabled {
  color: rgb(250, 247, 247);
}

.item-button.is-disabled:hover {
  color: rgb(0, 0, 0, 0.5);
  background: none;
}

.item-button:hover {
  background: rgb(134, 133, 133);
}

.content {
  padding: 4px 6px 4px 25px;
  color: white;
}
</style>
