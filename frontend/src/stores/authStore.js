import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { queryClient } from '@/lib/queryClient'

export const useAuthStore = create(
  persist(
    (set) => ({
      token: null,
      refreshToken: null,
      userId: null,
      username: null,
      login: (data) => {
        queryClient.clear()
        set({
          token: data.access_token,
          refreshToken: data.refresh_token,
          userId: data.user_id,
          username: data.username,
        })
      },
      setToken: (token) => set({ token }),
      setRefreshToken: (refreshToken) => set({ refreshToken }),
      logout: () => {
        queryClient.clear()
        set({ token: null, refreshToken: null, userId: null, username: null })
      },
    }),
    { name: 'ledger-auth' }
  )
)
