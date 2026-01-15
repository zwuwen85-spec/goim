import { defineStore } from 'pinia'
import { ref } from 'vue'
import { groupApi, messageApi, type Group, type GroupMember } from '../api/chat'

export const useGroupStore = defineStore('group', () => {
  const groups = ref<Group[]>([])
  const currentGroup = ref<Group | null>(null)
  const members = ref<GroupMember[]>([])
  const loading = ref(false)

  // Clear all data (for account switching)
  const clearAll = () => {
    groups.value = []
    currentGroup.value = null
    members.value = []
  }

  // Load user's groups
  const loadGroups = async () => {
    loading.value = true
    try {
      const response = await groupApi.getList()
      if ((response as any).code === 0) {
        // Filter out any undefined/null groups and ensure array
        const groupsData = (response as any).data?.groups || []
        groups.value = groupsData.filter((g: Group | null | undefined): g is Group => g != null && g.id != null)
      } else {
        groups.value = []
      }
    } catch (error: any) {
      console.error('Failed to load groups:', error)
      groups.value = []
      throw error
    } finally {
      loading.value = false
    }
  }

  // Create a new group
  const createGroup = async (data: { name: string; max_members?: number }): Promise<Group | null> => {
    loading.value = true
    try {
      const response = await groupApi.create(data)
      if ((response as any).code === 0) {
        const newGroup: Group = {
          id: (response as any).data.group_id,
          group_no: (response as any).data.group_no,
          name: (response as any).data.name,
          owner_id: (response as any).data.owner_id,
          max_members: (response as any).data.max_members,
          join_type: 1,
          mute_all: 0,
          member_count: 1,
          created_at: new Date().toISOString()
        }
        groups.value.push(newGroup)
        return newGroup
      }
      return null
    } catch (error: any) {
      console.error('Failed to create group:', error)
      throw error
    } finally {
      loading.value = false
    }
  }

  // Set current group
  const setCurrentGroup = async (group: Group) => {
    currentGroup.value = group
    await loadMembers(group.id)
  }

  // Load group members
  const loadMembers = async (groupId: number) => {
    try {
      const response = await groupApi.getMembers(groupId)
      if ((response as any).code === 0) {
        // Filter out any undefined/null members and ensure array
        const membersData = (response as any).data?.members || []
        members.value = membersData.filter((m: GroupMember | null | undefined): m is GroupMember => m != null && m.id != null)
      } else {
        members.value = []
      }
    } catch (error: any) {
      console.error('Failed to load members:', error)
      members.value = []
      throw error
    }
  }

  // Join a group
  const joinGroup = async (groupId: number, message?: string): Promise<boolean> => {
    try {
      const response = await groupApi.join(groupId, message)
      if ((response as any).code === 0) {
        await loadGroups()
        return true
      }
      return false
    } catch (error: any) {
      console.error('Failed to join group:', error)
      throw error
    }
  }

  // Leave a group
  const leaveGroup = async (groupId: number): Promise<boolean> => {
    try {
      const response = await groupApi.leave(groupId)
      if ((response as any).code === 0) {
        groups.value = groups.value.filter(g => g.id !== groupId)
        if (currentGroup.value?.id === groupId) {
          currentGroup.value = null
          members.value = []
        }
        return true
      }
      return false
    } catch (error: any) {
      console.error('Failed to leave group:', error)
      throw error
    }
  }

  // Send message to group
  const sendMessage = async (groupId: number, content: string): Promise<boolean> => {
    try {
      const response = await messageApi.send({
        to_group_id: groupId,
        conversation_type: 2, // group chat
        msg_type: 1, // text
        content: JSON.stringify({ text: content })
      })
      return (response as any).code === 0
    } catch (error: any) {
      console.error('Failed to send message:', error)
      throw error
    }
  }

  // Get group by ID
  const getGroupById = (groupId: number): Group | undefined => {
    return groups.value.find(g => g.id === groupId)
  }

  // Clear current group
  const clearCurrentGroup = () => {
    currentGroup.value = null
    members.value = []
  }

  return {
    groups,
    currentGroup,
    members,
    loading,
    clearAll,
    loadGroups,
    createGroup,
    setCurrentGroup,
    loadMembers,
    joinGroup,
    leaveGroup,
    sendMessage,
    getGroupById,
    clearCurrentGroup
  }
})
