# frozen_string_literal: true

require 'bundler/setup'
require 'discordrb'
require 'config'
require_relative './core/vcbot'
require_relative './core/voicevox'

# Config
config = Config.load_and_set_settings('./config.yml', './command.yml')

# init
bot = Discordrb::Commands::CommandBot.new(
  token: config.bot.token,
  client_id: config.bot.client_id,
  prefix: config.bot.prefix,
  ignore_bots: true
)
servers = []

# Bot Init
bot.ready do
  puts 'Bot is Ready!'
  bot.game = config.bot.status
end

# bot.include!(:help, { aliases: [:h], description: config.command.help.desc, usage: config.command.help.usage })

bot.command(:summon,
            { aliases: [:s], description: config.command.summon.desc, usage: config.command.summon.usage }) do |event|
  # コマンドを実行したユーザーがVCに接続しているか確認
  if event.user.voice_channel
    # サーバー用のThreadが起動しているか確認
    puts "Servers : #{servers.map(&:name)}"
    puts "ServerID: #{event.server.id}"
    if servers.none? { |sv| sv.name == event.server.id }
      server = VCBot.new(config, event)
      server.name = event.server.id
      event.respond('Hey!')
      servers << server
      puts "ServerInstance: #{server}"
      puts "Servers : #{servers.map(&:name)}"
      server.main
    else
      event.respond('すでにボイスチャットに接続されています。')
    end
  else
    event.respond('ボイスチャットに参加してから使用してください。')
  end
end

# bot.command(:stop, { aliases: [:skip] }) do |event|
#  event.voice.stop_playing if event.voice.playing?
# end

bot.command(:bye, { aliases: [:b], description: config.command.bye.desc, usage: config.command.bye.usage }) do |event|
  server = servers.find { |sv| sv.name == event.server.id }
  if server
    event.respond('Bye!')
    server.kill
    servers.delete(server)
  else
    event.respond('Botはボイスチャットに参加していません。')
  end
end

bot.command(:ping, { description: config.command.ping.desc, usage: config.command.ping.usage }) do |event|
  event.respond('Pong!')
end

bot.run
