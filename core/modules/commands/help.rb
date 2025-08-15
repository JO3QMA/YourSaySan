# frozen_string_literal: true

require 'yaml'

# Command Module
module YourSaySan
  module Commands
    # Help command
    module Help
      extend Discordrb::Commands::CommandContainer
      command(:help) do |event, command|
        config = YAML.load_file('command.yml')
        help_map = config['command'] || {}

        # 一覧表示
        if command.nil? || command.strip.empty?
          lines = [
            '読み上げBOTに使用できるコマンドです:',
            *help_map.map { |name, meta| "- !#{name}: #{meta['desc']}" }
          ]
          event.respond(lines.join("\n"))
          next
        end

        # 詳細表示
        meta = help_map[command]
        if meta
          event.respond("!#{command}: #{meta['desc']}\n使い方: #{meta['usage']}")
        else
          event.respond("'#{command}' は存在しません。!help で一覧を確認してください。")
        end
      end
    end
  end
end
