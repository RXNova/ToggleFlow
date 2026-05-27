import {
  TabsContent as TabsContentPrimitive,
  TabsList as TabsListPrimitive,
  TabsRoot,
  TabsTrigger as TabsTriggerPrimitive,
} from 'reka-ui'
import { cn } from '@/lib/utils'
import { h, defineComponent } from 'vue'

export const Tabs = TabsRoot

export const TabsList = defineComponent({
  name: 'TabsList',
  props: { class: String },
  setup(props, { slots }) {
    return () =>
      h(
        TabsListPrimitive,
        {
          class: cn(
            'inline-flex h-9 items-center rounded-lg bg-muted p-1 text-muted-foreground',
            props.class
          ),
        },
        slots
      )
  },
})

export const TabsTrigger = defineComponent({
  name: 'TabsTrigger',
  props: { class: String, value: { type: String, required: true as const } },
  setup(props, { slots }) {
    return () =>
      h(
        TabsTriggerPrimitive,
        {
          value: props.value,
          class: cn(
            'inline-flex items-center justify-center whitespace-nowrap rounded-md px-3 py-1 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 data-[state=active]:bg-background data-[state=active]:text-foreground data-[state=active]:shadow',
            props.class
          ),
        },
        slots
      )
  },
})

export const TabsContent = defineComponent({
  name: 'TabsContent',
  props: { class: String, value: { type: String, required: true as const } },
  setup(props, { slots }) {
    return () =>
      h(
        TabsContentPrimitive,
        {
          value: props.value,
          class: cn(
            'mt-2 ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
            props.class
          ),
        },
        slots
      )
  },
})
