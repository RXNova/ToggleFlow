<template>
  <DialogRoot :open="open" @update:open="$emit('update:open', $event)">
    <DialogPortal>
      <DialogOverlay
        class="fixed inset-0 z-50 bg-black/40 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0"
      />
      <DialogContent
        class="fixed right-0 top-0 z-50 flex h-full w-full max-w-xl flex-col border-l bg-background shadow-xl duration-300 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:slide-out-to-right data-[state=open]:slide-in-from-right focus:outline-none"
      >
        <!-- Header -->
        <div class="flex items-start justify-between border-b px-5 py-4 shrink-0">
          <div class="space-y-0.5">
            <DialogTitle class="text-sm font-semibold">
              {{ $t('rules.title') }}
            </DialogTitle>
            <DialogDescription class="text-xs text-muted-foreground">
              {{ flag.name }} · {{ envState.environment_name }}
            </DialogDescription>
          </div>
          <DialogClose
            class="rounded-sm p-1 opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
          >
            <X class="size-4" />
          </DialogClose>
        </div>

        <!-- Body -->
        <div class="flex min-h-0 flex-1 flex-col overflow-y-auto px-5 py-4 space-y-3">
          <p class="text-xs text-muted-foreground">{{ $t('rules.description') }}</p>

          <!-- Rule list -->
          <div
            v-if="rules.length === 0"
            class="flex flex-col items-center justify-center py-10 text-center"
          >
            <ListFilter class="size-8 text-muted-foreground/30 mb-2" />
            <p class="text-sm text-muted-foreground">{{ $t('rules.empty') }}</p>
          </div>

          <div v-else class="space-y-3">
            <div
              v-for="(rule, ri) in rules"
              :key="ri"
              class="rounded-lg border bg-card p-3 space-y-2"
            >
              <!-- Rule header -->
              <div class="flex items-center justify-between">
                <span class="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                  {{ $t('rules.rule') }} {{ ri + 1 }}
                </span>
                <button
                  class="rounded-md p-1 text-muted-foreground hover:bg-destructive/10 hover:text-destructive transition-colors"
                  @click="removeRule(ri)"
                >
                  <Trash2 class="size-3.5" />
                </button>
              </div>

              <!-- Conditions -->
              <div class="space-y-1.5">
                <div
                  v-for="(cond, ci) in rule.conditions"
                  :key="ci"
                  class="flex items-start gap-1.5"
                >
                  <span
                    v-if="ci > 0"
                    class="mt-1.5 shrink-0 text-[10px] font-semibold uppercase text-muted-foreground w-7 text-center"
                    >AND</span
                  >
                  <span v-else class="w-7 shrink-0" />

                  <div class="flex-1 grid grid-cols-[1fr_auto_1fr_auto] gap-1 items-start min-w-0">
                    <!-- Attribute -->
                    <Input
                      v-model="cond.attribute"
                      :placeholder="$t('rules.attribute')"
                      class="h-7 text-xs"
                    />
                    <!-- Operator -->
                    <select
                      v-model="cond.operator"
                      class="h-7 rounded-md border border-input bg-background px-2 text-xs focus:outline-none focus:ring-2 focus:ring-ring"
                      @change="onOperatorChange(cond)"
                    >
                      <optgroup :label="$t('rules.opGroupString')">
                        <option value="equals">{{ $t('rules.opEquals') }}</option>
                        <option value="in">{{ $t('rules.opIn') }}</option>
                        <option value="notIn">{{ $t('rules.opNotIn') }}</option>
                        <option value="contains">{{ $t('rules.opContains') }}</option>
                        <option value="startsWith">{{ $t('rules.opStartsWith') }}</option>
                        <option value="endsWith">{{ $t('rules.opEndsWith') }}</option>
                      </optgroup>
                      <optgroup :label="$t('rules.opGroupNumeric')">
                        <option value="gt">&gt;</option>
                        <option value="gte">&gt;=</option>
                        <option value="lt">&lt;</option>
                        <option value="lte">&lt;=</option>
                      </optgroup>
                    </select>
                    <!-- Values -->
                    <div class="min-w-0 space-y-1">
                      <!-- in / notIn: mode toggle + tag input or segment picker -->
                      <template v-if="cond.operator === 'in' || cond.operator === 'notIn'">
                        <!-- mode toggle -->
                        <div class="flex rounded-md border border-input overflow-hidden h-7 w-fit">
                          <button
                            type="button"
                            class="px-2 text-[11px] transition-colors"
                            :class="
                              !cond.segment
                                ? 'bg-muted font-medium'
                                : 'text-muted-foreground hover:bg-muted/50'
                            "
                            @click="cond.segment = undefined"
                          >
                            {{ $t('rules.modeValues') }}
                          </button>
                          <button
                            type="button"
                            class="px-2 text-[11px] transition-colors border-l border-input"
                            :class="
                              cond.segment
                                ? 'bg-muted font-medium'
                                : 'text-muted-foreground hover:bg-muted/50'
                            "
                            @click="cond.segment = cond.segment || (segments[0]?.key ?? '')"
                          >
                            {{ $t('rules.modeSegment') }}
                          </button>
                        </div>
                        <!-- tag input -->
                        <div
                          v-if="!cond.segment"
                          class="flex flex-wrap gap-1 rounded-md border border-input bg-background px-2 py-1 min-h-7 focus-within:ring-2 focus-within:ring-ring cursor-text"
                          @click="focusTagInput(ri, ci)"
                        >
                          <span
                            v-for="(val, vi) in cond.values"
                            :key="vi"
                            class="inline-flex items-center gap-0.5 rounded bg-muted px-1.5 py-0.5 text-[11px]"
                          >
                            {{ val }}
                            <button
                              class="hover:text-destructive"
                              @click.stop="removeTag(cond, vi)"
                            >
                              <X class="size-2.5" />
                            </button>
                          </span>
                          <input
                            :ref="(el) => setTagInputRef(el, ri, ci)"
                            class="flex-1 min-w-16 text-xs bg-transparent outline-none"
                            :placeholder="
                              cond.values.length === 0 ? $t('rules.tagPlaceholder') : ''
                            "
                            @keydown.enter.prevent="addTag(cond, $event)"
                            @keydown.backspace="onTagBackspace(cond, $event)"
                            @keydown="
                              (e: KeyboardEvent) =>
                                e.key === ',' && (e.preventDefault(), addTag(cond, e))
                            "
                            @paste="(e: ClipboardEvent) => onTagPaste(cond, e)"
                          />
                        </div>
                        <!-- segment picker -->
                        <div v-else>
                          <select
                            v-model="cond.segment"
                            class="h-7 w-full rounded-md border border-input bg-background px-2 text-xs focus:outline-none focus:ring-2 focus:ring-ring"
                          >
                            <option v-if="segments.length === 0" value="" disabled>
                              {{ $t('segments.noSegments') }}
                            </option>
                            <option v-for="seg in segments" :key="seg.key" :value="seg.key">
                              {{ seg.name }} ({{ seg.values.length }})
                            </option>
                          </select>
                        </div>
                      </template>
                      <!-- Single value input -->
                      <Input
                        v-else
                        :model-value="String(cond.values[0] ?? '')"
                        :placeholder="$t('rules.value')"
                        class="h-7 text-xs"
                        @update:model-value="cond.values = [$event]"
                      />
                    </div>
                    <!-- Remove condition -->
                    <button
                      v-if="rule.conditions.length > 1"
                      class="mt-1 rounded p-0.5 text-muted-foreground hover:text-destructive transition-colors"
                      @click="removeCondition(rule, ci)"
                    >
                      <X class="size-3.5" />
                    </button>
                    <span v-else class="w-5 shrink-0" />
                  </div>
                </div>
              </div>

              <!-- Add condition -->
              <button
                class="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors"
                @click="addCondition(rule)"
              >
                <Plus class="size-3" />{{ $t('rules.addCondition') }}
              </button>

              <!-- Serve -->
              <div class="pt-1 border-t space-y-2">
                <div class="flex items-center gap-2">
                  <span class="text-xs text-muted-foreground shrink-0">{{
                    $t('rules.serve')
                  }}</span>
                  <div class="flex rounded-md border border-input overflow-hidden h-6">
                    <button
                      type="button"
                      class="px-2 text-[11px] transition-colors"
                      :class="
                        !isRollout(rule)
                          ? 'bg-muted font-medium'
                          : 'text-muted-foreground hover:bg-muted/50'
                      "
                      @click="setFixed(ri)"
                    >
                      {{ $t('rules.serveFixed') }}
                    </button>
                    <button
                      type="button"
                      class="px-2 text-[11px] transition-colors border-l border-input"
                      :class="
                        isRollout(rule)
                          ? 'bg-muted font-medium'
                          : 'text-muted-foreground hover:bg-muted/50'
                      "
                      @click="setRollout(ri)"
                    >
                      {{ $t('rules.serveRollout') }}
                    </button>
                  </div>
                </div>

                <!-- Fixed: single variation -->
                <select
                  v-if="!isRollout(rule)"
                  v-model.number="rule.serve"
                  class="h-7 w-full rounded-md border border-input bg-background px-2 text-xs focus:outline-none focus:ring-2 focus:ring-ring"
                >
                  <option v-for="(v, i) in flag.variations" :key="i" :value="i">
                    {{ v.name }}
                  </option>
                </select>

                <!-- Rollout: weight per variation -->
                <div v-else class="space-y-1">
                  <div v-for="(step, si) in rule.rollout" :key="si" class="flex items-center gap-2">
                    <span class="flex-1 truncate text-xs text-muted-foreground">
                      {{ flag.variations[step.variation]?.name }}
                    </span>
                    <input
                      v-model.number="step.weight"
                      type="number"
                      min="0"
                      max="100"
                      class="h-7 w-16 rounded-md border border-input bg-background px-2 text-right text-xs focus:outline-none focus:ring-2 focus:ring-ring"
                    />
                    <span class="w-3 text-xs text-muted-foreground">%</span>
                  </div>
                  <div class="flex justify-end pt-0.5">
                    <span
                      class="text-xs"
                      :class="
                        rolloutTotal(rule) === 100
                          ? 'text-green-600 dark:text-green-400'
                          : 'text-destructive'
                      "
                    >
                      {{ $t('rules.rolloutTotal') }}: {{ rolloutTotal(rule) }}%
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Add rule -->
          <button
            class="flex w-full items-center justify-center gap-1.5 rounded-lg border border-dashed py-2.5 text-xs text-muted-foreground hover:border-foreground/30 hover:text-foreground transition-colors"
            @click="addRule"
          >
            <Plus class="size-3.5" />{{ $t('rules.addRule') }}
          </button>
        </div>

        <!-- Footer -->
        <div class="flex items-center justify-between border-t px-5 py-3 shrink-0">
          <Alert v-if="error" variant="destructive" class="py-1.5 px-3 flex-1 mr-3">
            <AlertDescription class="text-xs">{{ error }}</AlertDescription>
          </Alert>
          <div class="flex gap-2 ml-auto">
            <Button variant="outline" size="sm" @click="$emit('update:open', false)">
              {{ $t('common.cancel') }}
            </Button>
            <Button size="sm" :disabled="saving" @click="save">
              <Loader2 v-if="saving" class="size-3.5 animate-spin" />
              {{ saving ? $t('rules.saving') : $t('rules.save') }}
            </Button>
          </div>
        </div>
      </DialogContent>
    </DialogPortal>
  </DialogRoot>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogOverlay,
  DialogPortal,
  DialogRoot,
  DialogTitle,
} from 'reka-ui'
import { X, Plus, Trash2, ListFilter, Loader2 } from '@lucide/vue'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  flagsApi,
  type Flag,
  type FlagEnvState,
  type Rule,
  type Condition,
  type RolloutStep,
} from '@/api/flags'
import { segmentsApi, type Segment } from '@/api/segments'
import { useAsyncAction } from '@/composables/useAsyncAction'

const props = defineProps<{
  open: boolean
  flag: Flag
  envState: FlagEnvState
  projectId: number
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  saved: [rules: Rule[]]
}>()

const { t } = useI18n()

const rules = ref<Rule[]>([])
const segments = ref<Segment[]>([])
const { loading: saving, error, run } = useAsyncAction()

// tagInputRefs[ruleIndex][conditionIndex] = input element
const tagInputRefs = ref<Record<string, HTMLInputElement | null>>({})

function tagKey(ri: number, ci: number) {
  return `${ri}-${ci}`
}

function setTagInputRef(el: unknown, ri: number, ci: number) {
  tagInputRefs.value[tagKey(ri, ci)] = el as HTMLInputElement | null
}

function focusTagInput(ri: number, ci: number) {
  tagInputRefs.value[tagKey(ri, ci)]?.focus()
}

watch(
  () => props.open,
  async (v) => {
    if (v) {
      rules.value = JSON.parse(JSON.stringify(props.envState.rules ?? []))
      error.value = ''
      segments.value = await segmentsApi.list(props.projectId).catch(() => [])
    }
  },
  { immediate: true }
)

function newCondition(): Condition {
  return { attribute: '', operator: 'in', values: [] }
}

function newRule(): Rule {
  return { conditions: [newCondition()], serve: 0, rollout: [] }
}

function isRollout(rule: Rule): boolean {
  return Array.isArray(rule.rollout) && rule.rollout.length > 0
}

function rolloutTotal(rule: Rule): number {
  return (rule.rollout ?? []).reduce((sum, s) => sum + (s.weight || 0), 0)
}

function setFixed(ri: number) {
  const rule = rules.value[ri]
  rule.serve = rule.serve ?? 0
  rule.rollout = []
}

function setRollout(ri: number) {
  const rule = rules.value[ri]
  const n = props.flag.variations.length
  const base = Math.floor(100 / n)
  const remainder = 100 - base * n
  rule.rollout = props.flag.variations.map(
    (_, i): RolloutStep => ({
      variation: i,
      weight: i === 0 ? base + remainder : base,
    })
  )
  rule.serve = null
}

function addRule() {
  rules.value.push(newRule())
}

function removeRule(ri: number) {
  rules.value.splice(ri, 1)
}

function addCondition(rule: Rule) {
  rule.conditions.push(newCondition())
}

function removeCondition(rule: Rule, ci: number) {
  rule.conditions.splice(ci, 1)
}

function onOperatorChange(cond: Condition) {
  const multi = cond.operator === 'in' || cond.operator === 'notIn'
  if (!multi) {
    cond.values = cond.values.length > 0 ? [cond.values[0]] : []
    cond.segment = undefined
  }
}

function addTag(cond: Condition, event: Event) {
  const input = event.target as HTMLInputElement
  const val = input.value.trim()
  if (val && !cond.values.includes(val)) cond.values.push(val)
  input.value = ''
}

function removeTag(cond: Condition, vi: number) {
  cond.values.splice(vi, 1)
}

function onTagBackspace(cond: Condition, event: KeyboardEvent) {
  const input = event.target as HTMLInputElement
  if (input.value === '' && cond.values.length > 0) {
    cond.values.pop()
  }
}

function onTagPaste(cond: Condition, e: ClipboardEvent) {
  const pasted = e.clipboardData?.getData('text') ?? ''
  if (!pasted.includes(',') && !pasted.includes('\n')) return
  e.preventDefault()
  const input = e.target as HTMLInputElement
  const parts = (input.value + pasted)
    .split(/[,\n]/)
    .map((s) => s.trim())
    .filter(Boolean)
  for (const part of parts) {
    if (!cond.values.includes(part)) cond.values.push(part)
  }
  input.value = ''
}

async function save() {
  error.value = ''
  for (const rule of rules.value) {
    if (isRollout(rule) && rolloutTotal(rule) !== 100) {
      error.value = t('rules.errorRolloutTotal')
      return
    }
    for (const cond of rule.conditions) {
      if (!cond.attribute.trim()) {
        error.value = t('rules.errorAttribute')
        return
      }
      const usesSegment = (cond.operator === 'in' || cond.operator === 'notIn') && cond.segment
      if (
        !usesSegment &&
        (cond.values.length === 0 ||
          (cond.values.length === 1 && String(cond.values[0]).trim() === ''))
      ) {
        error.value = t('rules.errorValue')
        return
      }
    }
  }
  await run(async () => {
    await flagsApi.saveRules(
      props.projectId,
      props.flag.key,
      props.envState.environment_id,
      rules.value
    )
    emit('saved', rules.value)
    emit('update:open', false)
  })
}
</script>
