<template>
  <div class="group-list-container">
    <div class="list-header">
      <div class="list-title">我的群组</div>
      <el-button type="primary" size="small" circle @click="showCreateDialog = true">
        <el-icon><Plus /></el-icon>
      </el-button>
    </div>

    <div class="group-list">
      <template v-if="groupStore.loading">
        <el-skeleton :rows="3" animated class="p-4" />
      </template>
      <template v-else-if="groupStore.groups && groupStore.groups.length > 0">
        <div
          v-for="group in groupStore.groups.filter(g => g != null)"
          :key="group.id"
          class="list-item"
          :class="{ active: currentGroupId === group.id }"
          @click="handleSelect(group)"
        >
          <el-avatar :size="48" :style="{ backgroundColor: getGroupColor(group.id) }" shape="square" class="item-avatar">
            <el-icon size="24" color="white"><ChatDotRound /></el-icon>
          </el-avatar>
          <div class="item-content">
            <div class="item-name">{{ group.name }}</div>
            <div class="item-meta">{{ group.member_count }} 成员</div>
          </div>
        </div>
      </template>
      <el-empty v-else description="暂无群组" :image-size="60" />
    </div>

    <!-- Create Group Dialog -->
    <el-dialog v-model="showCreateDialog" title="创建群组" width="400px" append-to-body>
      <el-form :model="createForm" label-width="80px">
        <el-form-item label="群名称">
          <el-input v-model="createForm.name" placeholder="请输入群名称" maxlength="50" />
        </el-form-item>
        <el-form-item label="群人数">
          <el-input-number v-model="createForm.max_members" :min="2" :max="500" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreateGroup" :disabled="!createForm.name.trim()">
          创建
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Plus, ChatDotRound } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useGroupStore } from '../store/group'
import type { Group } from '../api/chat'

const emit = defineEmits(['select'])
const groupStore = useGroupStore()

const showCreateDialog = ref(false)
const createForm = ref({
  name: '',
  max_members: 100
})

const currentGroupId = computed(() => groupStore.currentGroup?.id)

const groupColors: Record<number, string> = {}
const getGroupColor = (groupId: number) => {
  if (!groupColors[groupId]) {
    const colors = ['#F56C6C', '#E6A23C', '#67C23A', '#409EFF', '#909399']
    groupColors[groupId] = colors[groupId % colors.length]
  }
  return groupColors[groupId]
}

const handleSelect = (group: Group) => {
  emit('select', group)
}

const handleCreateGroup = async () => {
  if (!createForm.value.name.trim()) return

  try {
    const newGroup = await groupStore.createGroup({
      name: createForm.value.name,
      max_members: createForm.value.max_members
    })

    if (newGroup) {
      ElMessage.success('群组创建成功')
      showCreateDialog.value = false
      createForm.value.name = ''
      createForm.value.max_members = 100
      // Auto select new group
      handleSelect(newGroup)
    }
  } catch (error: any) {
    ElMessage.error(error.message || '创建失败')
  }
}

onMounted(() => {
  if (!groupStore.groups.length) {
    groupStore.loadGroups()
  }
})
</script>

<style scoped>
.group-list-container {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.list-header {
  padding: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.list-title {
  font-weight: 600;
  color: var(--text-primary);
}

.group-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.list-item {
  display: flex;
  padding: 12px;
  gap: 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  margin-bottom: 4px;
}

.list-item:hover {
  background-color: var(--bg-body);
}

.list-item.active {
  background-color: var(--primary-light);
}

.item-avatar {
  flex-shrink: 0;
  border-radius: 12px;
}

.item-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 4px;
}

.item-name {
  font-weight: 500;
  color: var(--text-primary);
  font-size: 14px;
}

.item-meta {
  font-size: 12px;
  color: var(--text-secondary);
}
</style>
