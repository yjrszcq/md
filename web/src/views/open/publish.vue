<template>
  <div class="page-open-publish">
    <div class="title-view">公开文档</div>
    <el-table class="table-view" ref="tableRef" :data="tableData" height="100%" stripe border v-loading="tableLoading">
      <el-table-column prop="name" label="序号" align="center" width="100">
        <template #default="scope">
          {{ (tableCondition.page.current - 1) * tableCondition.page.size + scope.$index + 1 }}
        </template>
      </el-table-column>
      <el-table-column prop="name" label="文档名称" align="center" :label-class-name="columnClass.name">
        <template #header="scope">
          <el-popover v-model:visible="namePopover" width="170" trigger="click" :hide-after="0" @hide="tablePopoverHide">
            <el-input v-model="tableCondition.condition.name" placeholder="文档名称筛选" clearable @clear="namePopover = false"></el-input>
            <template #reference>
              <div style="cursor: pointer">文档名称</div>
            </template>
          </el-popover>
        </template>
      </el-table-column>
      <el-table-column prop="type" label="文档类型" align="center" :label-class-name="columnClass.type">
        <template #header="scope">
          <el-popover v-model:visible="typePopover" width="170" trigger="click" :hide-after="0" @hide="tablePopoverHide">
            <el-select v-model="tableCondition.condition.type" placeholder="文档类型筛选" clearable @clear="typePopover = false">
              <el-option label="Markdown" value="md" />
              <el-option label="OpenAPI" value="openApi" />
            </el-select>
            <template #reference>
              <div style="cursor: pointer">文档类型</div>
            </template>
          </el-popover>
        </template>
        <template #default="scope"> {{ scope.row.type === "md" ? "Markdown" : "OpenAPI" }} </template>
      </el-table-column>
      <el-table-column prop="bookName" label="文集名称" align="center" :label-class-name="columnClass.bookName">
        <template #header="scope">
          <el-popover v-model:visible="bookNamePopover" width="170" trigger="click" :hide-after="0" @hide="tablePopoverHide">
            <el-input v-model="tableCondition.condition.bookName" placeholder="文集名称筛选" clearable @clear="bookNamePopover = false"></el-input>
            <template #reference>
              <div style="cursor: pointer">文集名称</div>
            </template>
          </el-popover>
        </template>
      </el-table-column>
      <el-table-column prop="username" label="作者" align="center" :label-class-name="columnClass.username">
        <template #header="scope">
          <el-popover v-model:visible="usernamePopover" width="170" trigger="click" :hide-after="0" @hide="tablePopoverHide">
            <el-input v-model="tableCondition.condition.username" placeholder="作者筛选" clearable @clear="usernamePopover = false"></el-input>
            <template #reference>
              <div style="cursor: pointer">作者</div>
            </template>
          </el-popover>
        </template>
      </el-table-column>
      <el-table-column prop="createTime" label="创建时间" align="center">
        <template #default="scope"> {{ formatTime(scope.row.createTime, "YYYY-MM-DD HH:mm:ss") }} </template>
      </el-table-column>
      <el-table-column prop="createTime" label="修改时间" align="center">
        <template #default="scope"> {{ formatTime(scope.row.updateTime, "YYYY-MM-DD HH:mm:ss") }} </template>
      </el-table-column>
      <el-table-column label="文档地址" align="center" width="160">
        <template #default="scope">
          <div class="doc-link-actions">
            <el-button class="link-btn" text @click="copyClick(scope.row.id)">复制</el-button>
            <el-button class="link-btn" text @click="hrefClick(scope.row.id)">跳转</el-button>
          </div>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      background
      layout="total, sizes, prev, pager, next, jumper"
      :pageSizes="[10, 20, 50, 100]"
      v-model:pageSize="tableCondition.page.size"
      v-model:currentPage="tableCondition.page.current"
      :total="tableTotal"
      @size-change="tablePageSizeChange"
      @current-change="tablePageCurrentChange"
    ></el-pagination>
  </div>
</template>

<script lang="ts" setup>
import { ref, Ref, nextTick, onMounted } from "vue";
import { ElTable, ElMessage } from "element-plus";
import OpenApi from "@/api/open";
import { formatTime } from "@/utils";
import copy from "copy-to-clipboard";

const hostUrl = ref(location.origin);
const tableCondition = ref({
  page: { current: 1, size: 100 },
  condition: { name: "", type: "", bookName: "", username: "" },
});
const columnClass = ref({
  name: "",
  type: "",
  bookName: "",
  username: "",
});
const lastCondition = ref("");
const tableData: Ref<DocPageResult[]> = ref([]);
const tableTotal = ref(0);
const tableLoading = ref(false);
const tableRef = ref<InstanceType<typeof ElTable>>();
const namePopover = ref(false);
const typePopover = ref(false);
const bookNamePopover = ref(false);
const usernamePopover = ref(false);

onMounted(() => {
  queryTableData();
});

/**
 * 查询表格数据
 */
const queryTableData = () => {
  lastCondition.value = JSON.stringify(tableCondition.value.condition);
  tableLoading.value = true;
  OpenApi.pageDoc(tableCondition.value)
    .then((res) => {
      tableData.value = res.data.records;
      tableTotal.value = res.data.total;
      nextTick(() => {
        tableRef.value?.setScrollTop(0);
      });
    })
    .finally(() => {
      tableLoading.value = false;
    });
};

/**
 * 表格自定义筛选
 */
const tablePopoverHide = () => {
  if (tableCondition.value.condition.name) {
    columnClass.value.name = "column-active";
  } else {
    columnClass.value.name = "";
  }
  if (tableCondition.value.condition.type) {
    columnClass.value.type = "column-active";
  } else {
    columnClass.value.type = "";
  }
  if (tableCondition.value.condition.bookName) {
    columnClass.value.bookName = "column-active";
  } else {
    columnClass.value.bookName = "";
  }
  if (tableCondition.value.condition.username) {
    columnClass.value.username = "column-active";
  } else {
    columnClass.value.username = "";
  }
  if (JSON.stringify(tableCondition.value.condition) === lastCondition.value) {
    return;
  }
  tableCondition.value.page.current = 1;
  queryTableData();
};

/**
 * 每页显示条数变化
 * @param size
 */
const tablePageSizeChange = (size: number) => {
  tableCondition.value.page.current = 1;
  tableCondition.value.page.size = size;
  queryTableData();
};

/**
 * 当前页码变化
 * @param current
 */
const tablePageCurrentChange = (current: number) => {
  tableCondition.value.page.current = current;
  queryTableData();
};

/**
 * 点击复制
 * @param id
 */
const copyClick = (id: string) => {
  let url = hostUrl.value + "/#/open/document?id=" + id;
  const result = copy(url);
  if (result) {
    ElMessage.success("发布地址已复制到剪切板");
  } else {
    ElMessage.error("复制到剪切板失败，地址：" + url);
  }
};

/**
 * 点击跳转
 * @param id
 */
const hrefClick = (id: string) => {
  let url = hostUrl.value + "/#/open/document?id=" + id;
  window.open(url, "_blank");
};
</script>

<style lang="scss">
.page-open-publish {
  height: 100%;
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  background-color: var(--bg-secondary);
  .title-view {
    margin: 10px 0;
    font-size: 18px;
    color: var(--text-secondary);
    font-weight: bold;
    user-select: none;
  }
  .el-button.is-text + .el-button.is-text {
    margin-left: 0;
  }
  .column-active {
    color: #0094c1;
  }

  .doc-link-actions {
    display: inline-flex;
    gap: 8px;

    :deep(.link-btn) {
      color: var(--text-secondary);
      background-color: transparent !important;
      padding: 0 6px;

      &:hover,
      &:focus,
      &:focus-visible {
        color: var(--el-color-primary, #409eff);
        background-color: transparent !important;
      }
    }
  }
}
</style>
