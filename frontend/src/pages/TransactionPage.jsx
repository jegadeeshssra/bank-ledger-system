import { useParams, useNavigate } from 'react-router-dom'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table'
import { useTransaction } from '@/hooks/useAccounts'

const OPERATION_COLORS = {
  DEPOSIT:      'bg-[--credit]/15 text-[--credit] border-[--credit]/30',
  WITHDRAW:     'bg-destructive/15 text-destructive border-destructive/30',
  TRANSFER_IN:  'bg-[--credit]/15 text-[--credit] border-[--credit]/30',
  TRANSFER_OUT: 'bg-destructive/15 text-destructive border-destructive/30',
}

export default function TransactionPage() {
  const { id, transactionId } = useParams()
  const navigate = useNavigate()
  const { data: entries = [], isLoading, isError, error } = useTransaction(id, transactionId)

  return (
    <div className="space-y-5 max-w-6xl">
      <button
        onClick={() => navigate(`/accounts/${id}/entries`)}
        className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground transition-colors"
      >
        <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18" />
        </svg>
        Back to entries
      </button>

      <div className="space-y-2">
        <h1 className="text-xl font-semibold text-foreground">Transaction details</h1>
        <p className="text-sm text-muted-foreground">
          Showing all ledger entries for transaction <span className="font-mono">{transactionId}</span>.
        </p>
      </div>

      {isError && (
        <div className="rounded-md bg-destructive/10 border border-destructive/20 px-4 py-3 text-sm text-destructive">
          {error?.message || 'Unable to load transaction.'}
        </div>
      )}

      <div className="rounded-xl border border-border overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow className="border-border bg-muted/30">
              <TableHead className="text-xs uppercase tracking-wider text-muted-foreground">Type</TableHead>
              <TableHead className="text-xs uppercase tracking-wider text-muted-foreground text-right">Credit</TableHead>
              <TableHead className="text-xs uppercase tracking-wider text-muted-foreground text-right">Debit</TableHead>
              <TableHead className="text-xs uppercase tracking-wider text-muted-foreground hidden sm:table-cell">Entry ID</TableHead>
              <TableHead className="text-xs uppercase tracking-wider text-muted-foreground hidden md:table-cell">Date</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading &&
              Array.from({ length: 4 }).map((_, index) => (
                <TableRow key={index} className="border-border">
                  <TableCell><Skeleton className="h-4 w-20" /></TableCell>
                  <TableCell className="text-right"><Skeleton className="h-4 w-16 ml-auto" /></TableCell>
                  <TableCell className="text-right"><Skeleton className="h-4 w-16 ml-auto" /></TableCell>
                  <TableCell className="hidden sm:table-cell"><Skeleton className="h-4 w-28" /></TableCell>
                  <TableCell className="hidden md:table-cell"><Skeleton className="h-4 w-24" /></TableCell>
                </TableRow>
              ))}

            {!isLoading && entries.length === 0 && (
              <TableRow className="border-border">
                <TableCell colSpan={5} className="py-14 text-center text-sm text-muted-foreground">
                  No transaction entries found.
                </TableCell>
              </TableRow>
            )}

            {!isLoading && entries.map((entry) => {
              const credit = parseFloat(entry.credit ?? '0')
              const debit = parseFloat(entry.debit ?? '0')
              const colorClass = OPERATION_COLORS[entry.operation_type] || 'bg-secondary/15 text-secondary border-secondary/40'

              return (
                <TableRow key={entry.id} className="border-border hover:bg-muted/20 transition-colors">
                  <TableCell>
                    <Badge variant="outline" className={`text-xs font-medium border ${colorClass}`}>
                      {entry.operation_type}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-right font-mono tabular-nums text-sm">
                    {credit > 0 ? `+${credit.toFixed(2)}` : '—'}
                  </TableCell>
                  <TableCell className="text-right font-mono tabular-nums text-sm">
                    {debit > 0 ? `-${debit.toFixed(2)}` : '—'}
                  </TableCell>
                  <TableCell className="hidden sm:table-cell font-mono text-xs text-muted-foreground truncate">
                    {entry.id}
                  </TableCell>
                  <TableCell className="hidden md:table-cell text-xs text-muted-foreground">
                    {new Date(entry.created_at).toLocaleString()}
                  </TableCell>
                </TableRow>
              )
            })}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}
