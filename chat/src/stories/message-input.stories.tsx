import type { Meta, StoryObj } from "@storybook/nextjs";
import "../app/globals.css";

import MessageInput from "../components/message-input";

const meta = {
  title: "Components/MessageInput",
  component: MessageInput,
  parameters: {
    // More on how to position stories at: https://storybook.js.org/docs/configure/story-layout
    layout: "centered",
  },
} satisfies Meta<typeof MessageInput>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    onSendMessage: () => {},
  },
};
