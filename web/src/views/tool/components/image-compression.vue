<template>
  <div class="page-image-compression">
    <el-form label-width="100px">
      <el-form-item label="上传图片">
        <el-button type="primary" @click="upload">上传</el-button>
      </el-form-item>
      <el-form-item label="最大宽度">
        <el-input v-model="maxWidth" placeholder="请输入最大宽度，不填写则保持原图宽度"></el-input>
      </el-form-item>
      <el-form-item label="最大高度">
        <el-input v-model="maxHeight" placeholder="请输入最大高度，不填写则保持原图高度"></el-input>
      </el-form-item>
      <el-form-item label="图片质量">
        <el-slider v-model="quality" :min="0.1" :max="1" :step="0.1" show-stops show-input />
      </el-form-item>
      <el-form-item label="原图大小">
        {{ originalSize }}
      </el-form-item>
      <el-form-item label="压缩后大小">
        {{ compressSize }}
      </el-form-item>
    </el-form>
    <div class="image-view">
      <img :src="compressBase64" />
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, Ref, watch } from "vue";
import { ElMessage } from "element-plus";
import { Upload } from "@/utils";
import Compressor from "compressorjs";

const maxWidth = ref("");
const maxHeight = ref("");
const quality = ref(0.8);
const originalSize = ref("--");
const compressSize = ref("--");
const compressBase64 = ref("");
const file: Ref<File | undefined> = ref();

watch(
  () => maxWidth.value,
  () => {
    compress();
  }
);

watch(
  () => maxHeight.value,
  () => {
    compress();
  }
);

watch(
  () => quality.value,
  () => {
    compress();
  }
);

/**
 * 上传图片
 */
const upload = () => {
  Upload.openFiles(false, Upload.InputAccept.uploadImage).then((fileList) => {
    file.value = fileList[0];
    originalSize.value = Upload.formatFileSize(file.value.size);
    compress();
  });
};

/**
 * 压缩图片
 */
const compress = () => {
  if (!file.value) {
    return;
  }
  new Compressor(file.value, {
    quality: quality.value,
    maxWidth: parseInt(maxWidth.value),
    maxHeight: parseInt(maxHeight.value),
    success(result) {
      compressSize.value = Upload.formatFileSize(result.size);
      Upload.readBlobToBase64(result)
        .then((res) => {
          compressBase64.value = res;
        })
        .catch((err) => {
          clear();
          console.error(err);
          ElMessage.error("图片压缩失败");
        });
    },
    error(err) {
      clear();
      console.error(err);
      ElMessage.error("图片压缩失败");
    },
  });
};

/**
 * 清空全部信息
 */
const clear = () => {
  originalSize.value = "--";
  compressSize.value = "--";
  compressBase64.value = "";
  file.value = undefined;
};
</script>

<style lang="scss">
.page-image-compression {
  display: flex;
  flex-direction: column;
  align-items: center;
  .el-form {
    width: 400px;
    max-width: 100%;
  }
  .image-view {
    width: calc(100% - 40px);
    display: flex;
    align-items: center;
    justify-content: center;
    height: 50vh;
    min-height: 200px;
    background: #eee;
    padding: 20px;
    border-radius: 6px;
    img {
      max-width: 100%;
      max-height: 100%;
    }
  }
}

// Dark theme
[data-theme="dark"] {
  .page-image-compression {
    .image-view {
      background: #2a2a2a;
    }
  }
}
</style>
