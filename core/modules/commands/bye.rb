# frozen_string_literal: true

module YourSaySan
  module Commands
    # Byeコマンドモジュール
    module Bye
      extend Discordrb::EventContainer

      # コマンド情報を定義
      COMMAND_INFO = {
        desc: '読み上げBotを切断します。',
        usage: '引数は必要ありません。'
      }.freeze

      def self.register_slash_command(bot)
        bot.register_application_command(:bye, COMMAND_INFO[:desc])
      end

      application_command :bye do |event|
        if event.user.voice_channel
          YourSaySan::BOT.voice_destroy(event.server)
          event.respond(content: 'ボイスチャットから切断しました。', ephemeral: true)
        else
          event.respond(content: 'ボイスチャットに参加していません。', ephemeral: true)
        end
      rescue StandardError => e
        puts "[Bot] 切断エラー: #{e.message}"
        event.respond(content: '切断処理に失敗しました。', ephemeral: true)
      end
    end
  end
end
