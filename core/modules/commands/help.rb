# frozen_string_literal: true

# Command Module
module YouSaySan
  module Commands
    # Help command
    module Help
      extend Discordrb::Commands::CommandContainer
      command(:help) do |event, command|
        case command
        when ''
          event.respond('読み上げBOTに使用できるコマンドです。')
        when 'ping'
          '`Pong!` を返します。BOTの死活確認用のコマンドです。引数は取りません。'
        when 'summon'
          'BOTを召喚します。このコマンドを使用したテキストチャンネルを読み上げます。'
        end
      end
    end
  end
end
