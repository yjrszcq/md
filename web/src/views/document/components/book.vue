<template>
  <div class="page-book" v-loading="bookLoading" :class="{ 'page-book-shrink': !isStretch }">
    <div class="mask-view" v-if="loading"></div>
    <el-popover v-if="!onlyPreview" :visible="addBookVisible" placement="bottom" trigger="click" width="200px">
      <el-input v-model="newBookName" placeholder="请输入文集名称" style="margin-right: 10px"></el-input>
      <div style="display: flex; margin-top: 8px; justify-content: flex-end">
        <el-button @click="addBookCancel" size="small">取消</el-button>
        <el-button @click="addBookSave" type="primary" size="small">确定</el-button>
      </div>
      <template #reference>
        <el-button class="create-button" type="warning" size="large" link :icon="Plus" @click="addBookVisible = true">创建文集</el-button>
      </template>
    </el-popover>
    <el-button v-else class="create-button" type="warning" size="large" link>文集选择</el-button>
    <el-scrollbar class="scroll-view">
      <div class="item-view" :class="currentBookId === item.id ? 'selected' : ''" v-for="item in books" :key="item.id" @click="bookClick(item)">
        <div class="update-view" v-if="updateBookId && updateBookId === item.id">
          <el-input v-model="updateBookName" placeholder="请输入文集名称"></el-input>
          <el-button style="margin-left: 12px" type="success" link :icon="CircleCheckFilled" @click="updateBookSave"></el-button>
          <el-button type="danger" link :icon="CircleCloseFilled" @click="updateBookCancel"></el-button>
        </div>
        <text-tip :content="item.name" v-else></text-tip>
        <el-dropdown trigger="click" v-if="!onlyPreview && item.id">
          <el-icon class="setting-button" @click.stop="() => {}" title="操作"><Tools /></el-icon>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item style="user-select: none" @click="updateBookClick(item)">修改文集</el-dropdown-item>
              <el-dropdown-item style="user-select: none" @click="deleteBookClick(item)">删除文集</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </el-scrollbar>
  </div>
</template>

<script lang="ts" setup>
import { ref, Ref, onMounted, watch } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Plus, Tools, CircleCheckFilled, CircleCloseFilled } from "@element-plus/icons-vue";
import TextTip from "@/components/text-tip";
import BookApi from "@/api/book";

const books: Ref<Book[]> = ref([]);
const bookLoading = ref(false);
const currentBookId = ref("");
const addBookVisible = ref(false);
const newBookName = ref("");
const updateBookId = ref("");
const updateBookName = ref("");

const emit = defineEmits<{ change: [bookId: string]; books: [bookList: Book[]] }>();

defineProps({
  onlyPreview: {
    type: Boolean,
    default: true,
  },
  isStretch: {
    type: Boolean,
    default: true,
  },
  loading: {
    type: Boolean,
    default: false,
  },
});

watch(currentBookId, (val) => {
  emit("change", val);
});

onMounted(() => {
  queryBooks();
});

/**
 * 查询文集列表
 */
const queryBooks = () => {
  addBookCancel();
  updateBookCancel();
  bookLoading.value = true;
  BookApi.list()
    .then((res) => {
      books.value = res.data;
      emit("books", res.data);
    })
    .finally(() => {
      bookLoading.value = false;
    });
};

/**
 * 点击文集
 */
const bookClick = (book: Book) => {
  if (updateBookId.value) {
    return;
  }
  currentBookId.value = book.id;
};

/**
 * 点击添加文集保存
 */
const addBookSave = () => {
  let name = String(newBookName.value).trim();
  if (!name) {
    ElMessage.warning("请填写文集名称");
    return;
  }
  bookLoading.value = true;
  BookApi.add({ id: "", name: name })
    .then(() => {
      ElMessage.success("创建成功");
      queryBooks();
    })
    .catch(() => {
      bookLoading.value = false;
    });
};

/**
 * 点击添加文集取消
 */
const addBookCancel = () => {
  addBookVisible.value = false;
  newBookName.value = "";
};

/**
 * 点击修改文集
 */
const updateBookClick = (book: Book) => {
  updateBookId.value = book.id;
  updateBookName.value = book.name;
};

/**
 * 点击修改文集保存
 */
const updateBookSave = () => {
  let name = String(updateBookName.value).trim();
  if (!name) {
    ElMessage.warning("请填写文集名称");
    return false;
  }
  bookLoading.value = true;
  BookApi.update({ id: updateBookId.value, name: name })
    .then(() => {
      ElMessage.success("修改成功");
      queryBooks();
    })
    .catch(() => {
      bookLoading.value = false;
    });
};

/**
 * 点击修改文集取消
 */
const updateBookCancel = () => {
  updateBookId.value = "";
  updateBookName.value = "";
};

/**
 * 点击删除文集
 */
const deleteBookClick = (book: Book) => {
  ElMessageBox.confirm("是否删除文集：" + book.name + "？", "提示", {
    confirmButtonText: "确定",
    cancelButtonText: "取消",
    type: "warning",
  }).then(() => {
    BookApi.delete(book.id).then(() => {
      ElMessage.success("删除成功");
      if (currentBookId.value === book.id) {
        currentBookId.value = "";
      }
      queryBooks();
    });
  });
};
</script>

<style lang="scss">
.page-book {
  height: 100%;
  min-width: 220px;
  width: 220px;
  background: var(--book-sidebar-bg);
  display: flex;
  flex-direction: column;
  overflow-x: hidden;
  transition: margin-left 0.3s;
  .mask-view {
    position: absolute;
    width: 100%;
    height: 100%;
    z-index: 1000;
  }
  .create-button {
    height: 60px;
    border-bottom: 1px solid var(--border-dark) !important;
  }
  .el-button--large [class*="el-icon"] + span {
    margin-left: 3px;
  }
  .scroll-view {
    color: var(--book-sidebar-text);
    font-size: 13px;
    .item-view {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 18px 15px;
      cursor: pointer;
      border-left: 3px var(--book-sidebar-border) solid;
      transition: 0.05s;
      border-bottom: 1px solid var(--border-dark);
      .update-view {
        display: flex;
        align-items: center;
      }
    }
    .item-view:hover {
      background: var(--book-sidebar-selected);
      border-left-color: var(--book-sidebar-selected);
    }
    .item-view.selected {
      background: var(--book-sidebar-selected);
      border-left-color: #e6a23c;
    }
    .setting-button {
      margin-left: 10px;
      color: var(--book-sidebar-text);
    }
    .setting-button:hover {
      color: var(--text-tertiary);
    }
  }
  .el-loading-mask {
    background: var(--loading-mask-book);
  }
}
.page-book-shrink {
  margin-left: -220px;
}
@media (max-width: 480px) {
  .page-book {
    min-width: 45%;
    width: 45%;
  }
  .page-book-shrink {
    margin-left: -45%;
  }
}
</style>
