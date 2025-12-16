<template>
  <div class="page-doc" v-loading="docLoading" :class="{ 'page-doc-shrink': !isStretch }">
    <div class="mask-view" v-if="docDisabled"></div>
    <el-popover v-if="!onlyPreview" :visible="addDocVisible" placement="bottom" trigger="click" width="240px">
      <el-input v-model="newDocName" placeholder="请输入文档名称" style="margin-right: 10px"></el-input>
      <el-radio-group v-model="newDocType" style="margin-top: 8px">
        <el-radio-button value="md">Markdown</el-radio-button>
        <el-radio-button value="openApi">OpenAPI</el-radio-button>
      </el-radio-group>
      <div style="display: flex; margin-top: 8px; justify-content: flex-end">
        <el-button @click="addDocCancel" size="small">取消</el-button>
        <el-button @click="addDocSave" type="primary" size="small">确定</el-button>
      </div>
      <template #reference>
        <el-button class="create-button" type="primary" size="large" link :icon="Plus" @click="addDocVisible = true">创建文档</el-button>
      </template>
    </el-popover>
    <el-button v-else class="create-button" type="primary" size="large" link>文档选择</el-button>
    <el-scrollbar class="scroll-view" ref="scrollRef">
      <div
        class="item-view"
        :class="docIdTemp === item.id || (!docIdTemp && currentDoc.id === item.id) ? 'selected' : ''"
        v-for="item in docs"
        :key="item.id"
        @click="docClick(item)"
      >
        <text-tip :content="item.name"></text-tip>
        <div class="sub-text">{{ formatTime(item.updateTime, "YYYY-MM-DD HH:mm:ss") }}</div>
        <div v-if="item.published" class="published-view" @click.stop="copyPublishedClick(item)" title="已发布"></div>
        <el-dropdown trigger="click" v-if="!onlyPreview && item.id">
          <el-icon class="setting-button" @click.stop="() => {}" title="操作"><Tools /></el-icon>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item style="user-select: none" @click="updateDocClick(item)">修改文档</el-dropdown-item>
              <el-dropdown-item style="user-select: none" @click="deleteDocClick(item)">删除文档</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </el-scrollbar>
    <el-dialog
      v-model="dialog.visible"
      :title="dialog.isAdd ? '创建文档' : '更新文档信息'"
      width="400px"
      :show-close="false"
      :before-close="dialogClose"
    >
      <el-form label-width="70px" size="large">
        <el-form-item label="文档名称">
          <el-input v-model="dialog.condition.name" placeholder="请输入文档名称" style="width: 100%"></el-input>
        </el-form-item>
        <el-form-item label="文档类型">
          <el-radio-group v-model="dialog.condition.type" :disabled="!dialog.isAdd">
            <el-radio-button value="md">Markdown</el-radio-button>
            <el-radio-button value="openApi">OpenAPI</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="公开发布">
          <el-switch v-model="dialog.condition.published" />
        </el-form-item>
        <el-form-item label="所属文集">
          <el-select v-model="dialog.condition.bookId" style="width: 100%">
            <el-option v-for="item in books" :key="item.id" :label="item.name" :value="item.id"></el-option>
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button :loading="dialog.loading" @click="dialogClose">取消</el-button>
          <el-button type="primary" :loading="dialog.loading" @click="dialogSave">保存</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script lang="ts" setup>
import { ref, Ref, onMounted, watch, PropType, nextTick } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Plus, Tools } from "@element-plus/icons-vue";
import TextTip from "@/components/text-tip";
import DocumentApi from "@/api/document";
import { formatTime } from "@/utils";
import copy from "copy-to-clipboard";
import crypto from "crypto-js";
import NProgress from "nprogress";
import "nprogress/nprogress.css";

const hostUrl = ref(location.origin);
const docs: Ref<Doc[]> = ref([]);
const docLoading = ref(false);
const docDisabled = ref(false);
const addDocVisible = ref(false);
const newDocName = ref("");
const newDocType = ref("md");
const docIdTemp = ref("");
const dialog = ref({
  isAdd: true,
  loading: false,
  visible: false,
  condition: {
    id: "",
    name: "",
    content: "",
    bookId: "",
    type: "md",
    published: false,
  },
});
const scrollRef = ref();

const emit = defineEmits<{
  change: [id: string, name: string, content: string, type: string, updateTime: string, noRender?: boolean];
  loading: [val: boolean];
}>();

const props = defineProps({
  onlyPreview: {
    type: Boolean,
    default: true,
  },
  isStretch: {
    type: Boolean,
    default: true,
  },
  currentBookId: {
    type: String,
    default: "",
  },
  currentDoc: {
    type: Object as PropType<CurrentDoc>,
    default: () => ({
      id: "",
      content: "",
      originMD5: "",
      updateTime: "",
    }),
  },
  books: {
    type: Array as PropType<Book[]>,
    default: () => [],
  },
});

watch(
  () => props.currentBookId,
  (val) => {
    queryDocs(val);
  }
);

watch(docLoading, (val) => {
  emit("loading", val);
});

onMounted(() => {
  queryDocs(props.currentBookId);
});

/**
 * 查询文档列表
 */
const queryDocs = (bookId: string) => {
  addDocCancel();
  docLoading.value = true;
  DocumentApi.list(bookId)
    .then((res) => {
      // 如果文档id相同，但更新时间不同，丢弃缓存内容
      for (let item of res.data) {
        if (item.id === props.currentDoc.id) {
          if (String(item.updateTime) !== props.currentDoc.updateTime) {
            emitDoc("", "", "", "", "");
          }
          break;
        }
      }
      docs.value = res.data;
      // 滚动到当前文档位置
      nextTick(() => {
        scrollRef.value.$el.getElementsByClassName("item-view selected")[0]?.scrollIntoView();
      });
    })
    .finally(() => {
      docLoading.value = false;
    });
};

/**
 * 校验文档变化
 */
const checkDocChange = () => {
  return new Promise((resolve, reject) => {
    if (props.currentDoc.originMD5 && crypto.MD5(props.currentDoc.content).toString() !== props.currentDoc.originMD5) {
      ElMessageBox.confirm("文档未保存，是否继续？", "提示", {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        type: "warning",
      })
        .then(() => {
          resolve(null);
        })
        .catch(() => {
          reject();
        });
    } else {
      resolve(null);
    }
  });
};

/**
 * 回调文档信息
 */
const emitDoc = (id: string, name: string, content: string, type: string, updateTime: string, noRender?: boolean) => {
  emit("change", id, name, content, type, updateTime, noRender);
};

/**
 * 点击文档
 */
const docClick = (doc: Doc) => {
  checkDocChange().then(() => {
    docIdTemp.value = doc.id;
    docDisabled.value = true;
    emit("loading", true);
    NProgress.start();
    DocumentApi.get(doc.id)
      .then((res) => {
        emitDoc(res.data.id, res.data.name, res.data.content, res.data.type!, String(res.data.updateTime));
      })
      .finally(() => {
        docIdTemp.value = "";
        docDisabled.value = false;
        emit("loading", false);
        NProgress.done();
      });
  });
};

/**
 * 点击添加文档保存
 */
const addDocSave = () => {
  let name = String(newDocName.value).trim();
  if (!name) {
    ElMessage.warning("请填写文档名称");
    return;
  }
  checkDocChange().then(() => {
    docLoading.value = true;
    DocumentApi.add({ id: "", name: name, content: "", type: newDocType.value, bookId: props.currentBookId })
      .then((res) => {
        ElMessage.success("创建成功");
        emitDoc(res.data.id, res.data.name, res.data.content, res.data.type!, String(res.data.updateTime));
        queryDocs(props.currentBookId);
      })
      .catch(() => {
        docLoading.value = false;
      });
  });
};

/**
 * 点击添加文档取消
 */
const addDocCancel = () => {
  addDocVisible.value = false;
  newDocName.value = "";
  newDocType.value = "md";
};

/**
 * 点击修改文档
 */
const updateDocClick = (doc: Doc) => {
  dialog.value.condition.id = doc.id;
  dialog.value.condition.name = doc.name;
  dialog.value.condition.content = "";
  dialog.value.condition.bookId = doc.bookId;
  dialog.value.condition.type = doc.type!;
  dialog.value.condition.published = doc.published!;
  dialog.value.isAdd = false;
  dialog.value.visible = true;
};

/**
 * 点击删除文档
 */
const deleteDocClick = (doc: Doc) => {
  ElMessageBox.confirm("是否删除文档：" + doc.name + "？", "提示", {
    confirmButtonText: "确定",
    cancelButtonText: "取消",
    type: "warning",
  }).then(() => {
    DocumentApi.delete(doc.id).then(() => {
      ElMessage.success("删除成功");
      if (props.currentDoc.id === doc.id) {
        emitDoc("", "", "", "", "");
      }
      queryDocs(props.currentBookId);
    });
  });
};

/**
 * 点击发布地址
 * @param doc
 */
const copyPublishedClick = (doc: Doc) => {
  let url = hostUrl.value + "/#/open/document?id=" + doc.id;
  const result = copy(url);
  if (result) {
    ElMessage.success("发布地址已复制到剪切板");
  } else {
    ElMessage.error("复制到剪切板失败，地址：" + url);
  }
};

/**
 * 弹窗关闭
 */
const dialogClose = () => {
  if (dialog.value.loading) {
    return;
  }
  dialog.value.condition.id = "";
  dialog.value.condition.name = "";
  dialog.value.condition.content = "";
  dialog.value.condition.bookId = "";
  dialog.value.condition.type = "md";
  dialog.value.condition.published = false;
  dialog.value.isAdd = true;
  dialog.value.visible = false;
};

/**
 * 弹窗保存
 */
const dialogSave = () => {
  if (dialog.value.isAdd) {
    // 新增文档
    docLoading.value = true;
    DocumentApi.add(dialog.value.condition)
      .then((res) => {
        ElMessage.success("创建成功");
        emitDoc(res.data.id, res.data.name, res.data.content, res.data.type!, String(res.data.updateTime));
        docLoading.value = false;
        dialogClose();
        queryDocs(props.currentBookId);
      })
      .catch(() => {
        docLoading.value = false;
      });
  } else {
    // 更新基本信息
    let name = String(dialog.value.condition.name).trim();
    if (!name) {
      ElMessage.warning("请填写文档名称");
      return;
    }
    dialog.value.condition.name = name;
    dialog.value.loading = true;
    DocumentApi.update(dialog.value.condition)
      .then(() => {
        ElMessage.success("更新成功");
        dialog.value.loading = false;
        dialogClose();
        queryDocs(props.currentBookId);
      })
      .catch(() => {
        dialog.value.loading = false;
      });
  }
};

/**
 * 保存文档
 */
const saveDoc = (content: string) => {
  if (props.currentDoc.id !== "") {
    // 更新文档内容
    docLoading.value = true;
    DocumentApi.updateContent({ id: props.currentDoc.id, name: "", content: content, bookId: "" })
      .then((res) => {
        ElMessage.success("保存成功");
        emitDoc(res.data.id, res.data.name, res.data.content, res.data.type!, String(res.data.updateTime), true);
        // 更新当前文档的更新时间
        for (let item of docs.value) {
          if (item.id === res.data.id) {
            item.updateTime = res.data.updateTime;
            break;
          }
        }
      })
      .finally(() => {
        docLoading.value = false;
      });
  } else {
    // 新增
    dialog.value.condition.id = "";
    dialog.value.condition.name = "";
    dialog.value.condition.content = content;
    dialog.value.condition.bookId = "";
    dialog.value.condition.type = "md";
    dialog.value.condition.published = false;
    dialog.value.isAdd = true;
    dialog.value.visible = true;
  }
};

defineExpose({ saveDoc });
</script>

<style lang="scss">
.page-doc {
  height: 100%;
  min-width: 260px;
  width: 260px;
  background: #fafafa;
  display: flex;
  flex-direction: column;
  overflow-x: hidden;
  position: relative;
  transition: margin-left 0.3s;
  .mask-view {
    position: absolute;
    width: 100%;
    height: 100%;
    z-index: 1000;
  }
  .create-button {
    height: 60px;
    border-bottom: 1px solid #e6e6e6 !important;
  }
  .el-button--large [class*="el-icon"] + span {
    margin-left: 3px;
  }
  .scroll-view {
    color: #595959;
    font-size: 13px;
    .item-view {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 18px 15px;
      cursor: pointer;
      border-left: 3px #fafafa solid;
      transition: 0.05s;
      border-bottom: 1px solid #eaeaea;
      position: relative;
      .update-view {
        display: flex;
        align-items: center;
      }
      .sub-text {
        position: absolute;
        font-size: 12px;
        bottom: 3px;
        right: 20px;
        color: #ccc;
      }
      .published-view {
        position: absolute;
        top: 0;
        right: 0;
        width: 0;
        height: 0;
        border-top: 20px solid skyblue;
        border-left: 20px solid transparent;
      }
    }
    .item-view:hover {
      background: #e6e6e6;
      border-left-color: #e6e6e6;
    }
    .item-view.selected {
      background: #e6e6e6;
      border-left-color: #0094c1;
    }
    .setting-button {
      margin-left: 10px;
      color: #595959;
    }
    .setting-button:hover {
      color: #777;
    }
  }
  .el-loading-mask {
    background: #fafafa;
  }
}
.page-doc-shrink {
  margin-left: -260px;
}
@media (max-width: 480px) {
  .page-doc {
    min-width: 55%;
    width: 55%;
  }
  .page-doc-shrink {
    margin-left: -55%;
  }
}
</style>
