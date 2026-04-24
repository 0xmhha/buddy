#!/usr/bin/env node
import { Command } from 'commander'
import { runHookWrap } from './cli/hook-wrap.js'
import type { HookEventName } from './schema/hook-event.js'

const program = new Command()
program
  .name('buddy')
  .description('Claude Code 옆에서 hook 신뢰성을 지켜주는 친구.')
  .version('0.0.1')

program
  .command('hook-wrap <hook-name> [original-command...]')
  .description(
    'Claude Code hook을 감싸 실행한다. stdin/stdout/stderr/exit code를 ' +
      '그대로 전달하며, 실행 결과를 buddy outbox에 기록한다.',
  )
  .option('--event <name>', '기본 event 이름 (stdin이 비었을 때 사용)', 'PreToolUse')
  .option('--db <path>', 'buddy DB 경로 (기본: ~/.buddy/buddy.db)')
  .option('--record-tool-args', 'tool_input을 outbox에 기록 (기본 off)', false)
  .option(
    '--tag <kv...>',
    'customTags 추가. 형식: key=value. 여러 번 지정 가능.',
  )
  .action(async (hookName: string, originalCommand: string[], opts) => {
    const customTags = parseTags(opts.tag)
    const result = await runHookWrap({
      hookName,
      command: originalCommand,
      fallbackEvent: opts.event as HookEventName,
      dbPath: opts.db,
      recordToolArgs: opts.recordToolArgs === true,
      customTags,
    })
    process.exit(result.exitCode)
  })

function parseTags(raw: string[] | undefined): Record<string, string> | undefined {
  if (!raw || raw.length === 0) return undefined
  const out: Record<string, string> = {}
  for (const item of raw) {
    const eq = item.indexOf('=')
    if (eq <= 0) continue
    const key = item.slice(0, eq).trim()
    const value = item.slice(eq + 1).trim()
    if (key) out[key] = value
  }
  return Object.keys(out).length > 0 ? out : undefined
}

program.parseAsync(process.argv).catch((err) => {
  process.stderr.write(`buddy: ${(err as Error).message ?? 'fatal'}\n`)
  process.exit(2)
})
