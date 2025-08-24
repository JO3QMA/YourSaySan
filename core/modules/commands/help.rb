# frozen_string_literal: true

require 'yaml'

# Command Module
module YourSaySan
  module Commands
    # Help command
    module Help
      extend Discordrb::EventContainer

      # コマンド情報を定義
      COMMAND_INFO = {
        desc: 'ヘルプを表示します。',
        usage: '`/help [command]` で使用方法を表示します。'
      }.freeze

      def self.register_slash_command(bot)
        bot.register_application_command(:help, COMMAND_INFO[:desc]) do |cmd|
          cmd.string(:command, 'コマンド名（省略可）', required: false)
        end
      end

      # 全コマンドの情報を自動取得
      def self.get_all_commands_info
        commands_info = {}

        # Commandsモジュール内の全ての定数を取得
        YourSaySan::Commands.constants.each do |const_name|
          const = YourSaySan::Commands.const_get(const_name)

          # モジュールで、COMMAND_INFO定数を持っているものを対象とする
          next unless const.is_a?(Module) && const.const_defined?(:COMMAND_INFO)

          # コマンド名を取得（モジュール名をスネークケースに変換）
          command_name = const_name.to_s.downcase

          # COMMAND_INFOを取得
          command_info = const.const_get(:COMMAND_INFO)

          # ハッシュ形式で保存
          commands_info[command_name] = {
            'desc' => command_info[:desc],
            'usage' => command_info[:usage]
          }
        end

        commands_info
      end

      application_command(:help) do |event|
        help_map = get_all_commands_info

        # オプションを取得（複数の方法を試す）
        command_option = nil
        if event.respond_to?(:options) && event.options.is_a?(Hash)
          command_option = event.options['command']
        elsif event.respond_to?(:data) && event.data&.options
          command_option = event.data.options.find { |opt| opt.name == 'command' }&.value
        end

        # 一覧表示
        if command_option.nil? || command_option.to_s.strip.empty?
          lines = [
            '読み上げBOTに使用できるコマンドです:',
            *help_map.map { |name, meta| "- /#{name}: #{meta['desc']}" }
          ]
          event.respond(content: lines.join("\n"), ephemeral: true)
          next
        end

        # 詳細表示
        meta = help_map[command_option.to_s]
        if meta
          event.respond(content: "/#{command_option}: #{meta['desc']}\n使い方: #{meta['usage']}", ephemeral: true)
        else
          event.respond(content: "'#{command_option}' は存在しません。/help で一覧を確認してください。", ephemeral: true)
        end
      end
    end
  end
end
