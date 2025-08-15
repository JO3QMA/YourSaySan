# frozen_string_literal: true

module YourSaySan
  module Commands
    # Evalコマンドモジュール
    module Eval
      extend Discordrb::EventContainer

      def self.register_slash_command(bot)
        bot.register_application_command(:eval, 'コードを実行します（開発者用）') do |cmd|
          cmd.string(:code, '実行するコード', required: true)
        end
      end

      application_command :eval do |event|
        return unless event.user.id == YourSaySan::CONFIG.bot.owner

        # オプションを取得（複数の方法を試す）
        code = nil
        if event.respond_to?(:options) && event.options.is_a?(Hash)
          code = event.options['code']
        elsif event.respond_to?(:data) && event.data&.options
          code = event.data.options.find { |opt| opt.name == 'code' }&.value
        end
        
        return event.respond(content: 'コードが指定されていません。', ephemeral: true) unless code

        begin
          result = eval code
          event.respond(content: "実行結果: #{result.to_s}", ephemeral: true)
        rescue StandardError => e
          event.respond(content: "エラー: #{e.message}", ephemeral: true)
        end
      end
    end
  end
end
