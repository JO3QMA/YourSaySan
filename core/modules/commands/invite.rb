# frozen_string_literal: true

module YourSaySan
  module Commands
    # Inviteコマンドモジュール
    module Invite
      extend Discordrb::EventContainer

      COMMAND_INFO = {
        desc: 'Botを他のサーバーに招待するためのURLを表示します',
        usage: '`/invite` でBotを他のサーバーに招待するためのURLを表示します。'
      }.freeze

      def self.register_slash_command(bot)
        bot.register_application_command(:invite, 'Botを他のサーバーに招待するためのURLを表示します')
      end

      application_command :invite do |event|
        client_id = YourSaySan::CONFIG.bot.client_id

        # Botに必要な権限を設定
        # 読み上げBotに必要な権限:
        # - メッセージを送信 (Send Messages): 2048
        # - 音声チャンネルに接続 (Connect): 1048576
        # - 音声を再生 (Speak): 2097152
        # - チャンネルを見る (View Channels): 1024
        permissions = 2048 + 1_048_576 + 2_097_152 + 1024

        invite_url = "https://discord.com/api/oauth2/authorize?client_id=#{client_id}&permissions=#{permissions}&scope=bot"

        embed = Discordrb::Webhooks::Embed.new(
          title: '🤖 Bot招待リンク',
          description: '以下のリンクからBotを他のサーバーに招待できます！',
          color: 0x00ff00,
          fields: [
            Discordrb::Webhooks::EmbedField.new(
              name: '招待URL',
              value: invite_url,
              inline: false
            ),
            Discordrb::Webhooks::EmbedField.new(
              name: '必要な権限',
              value: "• メッセージを送信\n• 音声チャンネルに接続\n• 音声を再生\n• チャンネルを見る",
              inline: false
            )
          ],
          footer: Discordrb::Webhooks::EmbedFooter.new(text: '読み上げBot招待リンク')
        )

        event.respond(embeds: [embed], ephemeral: true)
      end
    end
  end
end
