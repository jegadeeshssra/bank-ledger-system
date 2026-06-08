import { useParams, useNavigate, Link } from 'react-router-dom'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table'
import { useEntries, useAccount } from '@/hooks/useAccounts'

const OPERATION_COLORS = {
  DEPOSIT:      'bg-emerald-100 text-emerald-700 border-emerald-200',
  WITHDRAW:     'bg-rose-100 text-rose-700 border-rose-200',
  TRANSFER_IN:  'bg-emerald-100 text-emerald-700 border-emerald-200',
  TRANSFER_OUT: 'bg-rose-100 text-rose-700 border-rose-200',
}

export default function EntriesPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { data: account } = useAccount(id)
  const { data: entries, isLoading, isError, error } = useEntries(id)

  return (
    <div className="space-y-5 max-w-6xl">
      {/* Back */}
      <button
        onClick={() => navigate(`/accounts/${id}`)}
        className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground transition-colors"
      >
        <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18" />
        </svg>
        Back to account
      </button>

      {/* Header */}
      <div>
        <h1 className="text-xl font-semibold text-foreground">
          {account?.name ? `${account.name} — Entries` : 'Entries'}
        </h1>
        <p className="text-sm text-muted-foreground mt-0.5">
          {isLoading ? 'Loading…' : `${entries?.length ?? 0} entries`}
        </p>
      </div>

      {/* Error */}
      {isError && (
        <div className="rounded-md bg-destructive/10 border border-destructive/20 px-4 py-3 text-sm text-destructive">
          {error?.message || 'Failed to load entries.'}
        </div>
      )}

      {/* Table */}
      <div className="rounded-xl border border-border overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow className="border-border bg-muted/30 hover:bg-muted/30">
              <TableHead className="text-xs uppercase tracking-wider text-muted-foreground w-28">Type</TableHead>
              <TableHead className="text-xs uppercase tracking-wider text-muted-foreground text-right">Credit</TableHead>
              <TableHead className="text-xs uppercase tracking-wider text-muted-foreground text-right">Debit</TableHead>
              <TableHead className="text-xs uppercase tracking-wider text-muted-foreground hidden sm:table-cell">Transaction</TableHead>
              <TableHead className="text-xs uppercase tracking-wider text-muted-foreground hidden md:table-cell">Date</TableHead>
              <TableHead className="w-8" />
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading &&
              Array.from({ length: 5 }).map((_, i) => (
                <TableRow key={i} className="border-border">
                  <TableCell><Skeleton className="h-5 w-24" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-16 ml-auto" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-16 ml-auto" /></TableCell>
                  <TableCell className="hidden sm:table-cell"><Skeleton className="h-4 w-24" /></TableCell>
                  <TableCell className="hidden md:table-cell"><Skeleton className="h-4 w-20" /></TableCell>
                  <TableCell />
                </TableRow>
              ))}

            {!isLoading && entries?.length === 0 && (
              <TableRow className="border-border">
                <TableCell colSpan={6} className="py-12 text-center text-sm text-muted-foreground">
                  No entries found for this account
                </TableCell>
              </TableRow>
            )}

            {!isLoading &&
              entries?.map((entry) => {
                const credit = parseFloat(entry.credit ?? '0')
                const debit = parseFloat(entry.debit ?? '0')
                const opColor = OPERATION_COLORS[entry.operation_type] || 'bg-secondary text-secondary-foreground border-border'

                return (
                  <TableRow key={entry.id} className="border-border hover:bg-muted/20 transition-colors">
                    <TableCell>
                      <Badge variant="outline" className={`text-xs font-medium border ${opColor}`}>
                        {entry.operation_type}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-right font-mono tabular-nums text-sm">
                      {credit > 0 ? (
                        <span className="text-[--credit]">+{credit.toFixed(2)}</span>
                      ) : (
                        <span className="text-muted-foreground/50">—</span>
                      )}
                    </TableCell>
                    <TableCell className="text-right font-mono tabular-nums text-sm">
                      {debit > 0 ? (
                        <span className="text-destructive">−{debit.toFixed(2)}</span>
                      ) : (
                        <span className="text-muted-foreground/50">—</span>
                      )}
                    </TableCell>
                    <TableCell className="hidden sm:table-cell">
                      <Link
                        to={`/accounts/${id}/transactions/${entry.transaction_id}`}
                        className="text-xs font-mono text-primary hover:underline"
                        onClick={(e) => e.stopPropagation()}
                      >
                        {entry.transaction_id.slice(0, 8)}…
                      </Link>
                    </TableCell>
                    <TableCell className="hidden md:table-cell text-xs text-muted-foreground">
                      {new Date(entry.created_at).toLocaleString()}
                    </TableCell>
                    <TableCell>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="w-7 h-7 text-muted-foreground hover:text-foreground"
                        onClick={() => navigate(`/accounts/${id}/transactions/${entry.transaction_id}`)}
                      >
                        <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                          <path strokeLinecap="round" strokeLinejoin="round" d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3" />
                        </svg>
                      </Button>
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
