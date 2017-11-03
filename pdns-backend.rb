#! /usr/bin/ruby
# -*- coding: utf-8 -*-

require 'json'
require 'pp'
require 'syslog'


class LIST
  attr_accessor :ip, :weight, :ttl

  def initialize(ip, weight, ttl)
    @ip = ip
    @weight = weight
    @ttl = ttl
  end
end


def random_choice(dup)
  dup.sort{|a,b| a.weight <=> b.weight}
  weight_sum = 0
  dup.each do |d|
    weight_sum +=  d.weight
  end
  dup.each do |d|
    d.weight /= weight_sum
  end
  r = rand()
  sum = 0
  dup.each_with_index do |d, n|
    sum += d.weight
    return n if r < sum
  end
  return dup.length - 1
end


def sort(ip,select_num)
  lis1 = Array.new
  lis2 = Array.new
  cnt = 0

  # ipアドレスと重みを配列に格納
  ip.each_with_index do |i, n|
 #   s = i.split(" ")
    lis1[n] = LIST.new(i["ip"], i["weight"].to_f, i["ttl"].to_i) 
  end
  
  # 重み付けしてソート
  loop{
    # すべてのIPの重みが0の場合
    if !checkweight(lis1) then
      if lis2.length == 0 then
        return lis1
      end
      break
    end
    if lis1.length == 1 && cnt == 0 then
      lis2.push(lis1[0])
        break
    end
    if lis1.length == 1 then
      #lis2.push(lis1[0])
#        break
    end
    break if cnt >= select_num.to_i # いくつIPアドレスを返すか
    index = random_choice(lis1)
    lis2.push(lis1[index])
    lis1.delete_at(index)
    cnt += 1
  }
  return lis2
end

def checkhash(qname, qtype)

  qtype = "A" if qtype == "ANY"
 
  return false if @@config["domain"] != qname
  return false if @@config["type"] != qtype.downcase
  return true
  
end

def checkweight(lis)

  lis.each do |i|
    if i.weight != 0 then
      return true
    end
  end
  return false
end


def minimum_ttl(list)
  ttl = 1000
  list.each do |i|
    if i.ttl < ttl
      ttl = i.ttl
    end
  end
  
  return ttl
  
end


$stdout.sync = true
$syslog = Syslog.open(__FILE__)
END { $syslog.close }

line = gets
line.strip!

unless line == "HELO\t1"
  puts "FAIL"
  #$syslog.err "Recevied '#{line}'"
  gets
  exit
end

puts "OK Sample backend firing up\t"

while gets
  #$syslog.info "#{$$} Received: #{$_}"
  $_.strip!
  arr = $_.split(/\t/)

  if (arr.length < 6)
    puts "LOG\tPowerDNS sent unparseable line"
    puts "FAIL"
    next
  end
  type, qname, qclass, qtype, id, ip = arr
  #$syslog.info "#{type}, #{qname}, #{qclass}, #{qtype}, #{id}"

  File.open("/home/vagrant/backend/wrr-config.json") do |file|
    @@config = JSON.load(file)
  end
  
  if checkhash(qname, qtype)
    if ["A", "ANY"].any? {|i| qtype == i } 
      arr = sort(@@config["record"], @@config["num"])
      ttl = minimum_ttl(arr)
      #$syslog.info "#{$$} Sent A records"
      arr.each do |i|
        #$syslog.info  ["send : ", "DATA", qname, qclass, "A", i.ttl, 1, i.ip].join(" ")
        puts ["DATA", qname, qclass, "A", ttl, 1, i.ip].join("\t")
      end
    end
  end

  #$syslog.info "#{$$} End of data"
  puts "END"

end

