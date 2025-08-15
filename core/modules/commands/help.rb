# frozen_string_literal: true

require 'yaml'

# Command Module
module YourSaySan
  module Commands
    # Help command
    module Help
      extend Discordrb::EventContainer
      
      def self.register_slash_command(bot)
        bot.register_application_command(:help, 'ヘルプを表示します') do |cmd|
          cmd.string(:command, 'コマンド名（省略可）', required: false)
        end
      end
      
      application_command(:help) do |event|
        config = YAML.load_file('command.yml')
        help_map = config['command'] || {}
        
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
