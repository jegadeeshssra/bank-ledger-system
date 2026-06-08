import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import * as api from '@/lib/api'

// ── Queries ───────────────────────────────────────────────────────────

export function useAccounts() {
  return useQuery({
    queryKey: ['accounts'],
    queryFn: api.listAccounts,
  })
}

export function useAccount(id) {
  return useQuery({
    queryKey: ['accounts', id],
    queryFn: () => api.getAccount(id),
    enabled: !!id,
  })
}

export function useEntries(accountId) {
  return useQuery({
    queryKey: ['entries', accountId],
    queryFn: () => api.getEntries(accountId),
    enabled: !!accountId,
  })
}

export function useTransaction(accountId, transactionId) {
  return useQuery({
    queryKey: ['transaction', accountId, transactionId],
    queryFn: () => api.getTransaction(accountId, transactionId),
    enabled: !!(accountId && transactionId),
  })
}

// ── Mutations ─────────────────────────────────────────────────────────

export function useCreateAccount(onSuccessCallback) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: api.createAccount,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['accounts'] })
      toast.success('Account created successfully')
      onSuccessCallback?.()
    },
    onError: (err) => toast.error(err.message),
  })
}

export function useDeleteAccount() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: api.deleteAccount,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['accounts'] })
      toast.success('Account deleted')
    },
    onError: (err) => toast.error(err.message),
  })
}

export function useDeposit(accountId, onSuccessCallback) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ amount, idempotencyKey }) => api.deposit(accountId, amount, idempotencyKey),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['accounts', accountId] })
      qc.invalidateQueries({ queryKey: ['accounts'] })
      qc.invalidateQueries({ queryKey: ['entries', accountId] })
      toast.success('Deposit successful')
      onSuccessCallback?.()
    },
    onError: () => toast.error('Deposit failed. Please try again.'),
  })
}

export function useWithdraw(accountId, onSuccessCallback) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ amount, idempotencyKey }) => api.withdraw(accountId, amount, idempotencyKey),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['accounts', accountId] })
      qc.invalidateQueries({ queryKey: ['accounts'] })
      qc.invalidateQueries({ queryKey: ['entries', accountId] })
      toast.success('Withdrawal successful')
      onSuccessCallback?.()
    },
    onError: () => toast.error('Withdrawal failed. Please try again.'),
  })
}

export function useTransfer(fromAccountId, onSuccessCallback) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ data, idempotencyKey }) => api.transfer(fromAccountId, data, idempotencyKey),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['accounts'] })
      qc.invalidateQueries({ queryKey: ['entries'] })
      toast.success('Transfer successful')
      onSuccessCallback?.()
    },
    onError: () => toast.error('Transfer failed. Please try again.'),
  })
}

export function useReconcile(accountId, onSuccessCallback) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => api.reconcile(accountId),
    onSuccess: (data) => {
      qc.invalidateQueries({ queryKey: ['accounts', accountId] })
      qc.invalidateQueries({ queryKey: ['accounts'] })
      toast.success(`Reconciled — Balance: ${data?.balance ?? 'updated'}`)
      onSuccessCallback?.()
    },
    onError: (err) => toast.error(err.message),
  })
}
